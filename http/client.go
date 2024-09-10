package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	party "github.com/h4lim/client-party"
	"github.com/h4lim/og-kds/infra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ClientRequest struct {
	HttpMethod       string
	URL              string
	Header           map[string]string
	RequestBody      *string
	QueryParam       *map[string]string
	Username         *string
	Password         *string
	AdditionalTracer *[]string
	ResponseId       int64
}

type ClientContext struct {
	ClientRequest ClientRequest
	PartyResponse *party.Response
	Error         error
	DataOptConfig OptConfig
}

type IClient interface {
	Hit() ClientContext
	MustHttpOk200() ClientContext
	UnmarshalJson(jsonData any) ClientContext
	GetPartyResponse() (*party.Response, error)
}

func NewClient(ctx ClientRequest, optConfig ...OptConfig) IClient {
	var _optConfig OptConfig
	if len(optConfig) > 0 {
		_optConfig = optConfig[len(optConfig)-1]
	}

	return ClientContext{
		ClientRequest: ctx,
		DataOptConfig: _optConfig,
	}
}

// GetPartyResponse implements IClient.
func (c ClientContext) GetPartyResponse() (*party.Response, error) {
	if c.Error != nil {
		return nil, c.Error
	}

	return c.PartyResponse, nil
}

// MustHttpOk200 implements IClient.
func (c ClientContext) MustHttpOk200() ClientContext {
	if c.Error == nil {
		if c.PartyResponse.HttpCode != 200 {
			c.Error = errors.New("http must 200 : " +
				strconv.Itoa(c.PartyResponse.HttpCode) +
				"response body" + c.PartyResponse.ResponseBody)
		}
	}

	return c
}

// UnmarshalJson implements IClient.
func (c ClientContext) UnmarshalJson(jsonData any) ClientContext {
	if c.Error == nil {
		errUnmarshal := json.Unmarshal([]byte(c.PartyResponse.ResponseBody), jsonData)
		if errUnmarshal != nil {
			c.Error = errUnmarshal
		}
	}

	return c
}

func (c ClientContext) Hit() ClientContext {

	clientParty := party.NewClientParty(c.ClientRequest.HttpMethod, c.ClientRequest.URL).
		SetHeader(c.ClientRequest.Header["Content-Type"], c.ClientRequest.Header)

	var zapFields []zapcore.Field

	if c.ClientRequest.AdditionalTracer != nil {
		zapFields = append(zapFields, zap.String("additional-tracer", strings.Join(*c.ClientRequest.AdditionalTracer, " ")))
	}

	zapFields = append(zapFields, zap.String("url", c.ClientRequest.URL))
	zapFields = append(zapFields, zap.String("http-methode", c.ClientRequest.HttpMethod))
	zapFields = append(zapFields, zap.String("header", fmt.Sprintf("%v", &c.ClientRequest.Header)))

	if c.ClientRequest.RequestBody != nil {
		clientParty.SetRequestBodyStr(*c.ClientRequest.RequestBody)
		zapFields = append(zapFields, zap.String("request-body", *c.ClientRequest.RequestBody))
	}

	if c.ClientRequest.QueryParam != nil {
		clientParty.SetQueryParam(*c.ClientRequest.QueryParam)
		zapFields = append(zapFields, zap.String("query-param",
			fmt.Sprintf("%v", *c.ClientRequest.QueryParam)))
	}

	if c.ClientRequest.Username != nil && c.ClientRequest.Password != nil {
		clientParty.SetBaseAuth(*c.ClientRequest.Username, *c.ClientRequest.Password)
	}

	clientResponse, err := clientParty.HitClient()
	if err != nil {
		zapFields = append(zapFields, zap.String("query-param",
			fmt.Sprintf("%v", *err)))
		c.Error = *err
	}

	zapFields = append(zapFields, zap.String("step", GetNextStep(c.ClientRequest.ResponseId)))

	duration := GetDuration(c.ClientRequest.ResponseId) + " ms"
	zapFields = append(zapFields, zap.String("duration", duration))

	totalDuration := time.Now().UnixNano() - c.ClientRequest.ResponseId
	ms := float64(totalDuration) / float64(time.Millisecond)
	zapFields = append(zapFields, zap.String("total-duration", fmt.Sprintf("%v", ms)+" ms"))

	zapFields = append(zapFields, zap.Int("http-code",
		clientResponse.HttpCode))

	zapFields = append(zapFields, zap.String("client-response-body",
		clientResponse.ResponseBody))

	zapFields = append(zapFields, zap.String("client-response-header",
		fmt.Sprintf("%v", clientResponse.ResponseHeader)))

	if infra.ZapLog != nil {
		infra.ZapLog.Debug(strconv.FormatInt(c.ClientRequest.ResponseId, 10), zapFields...)
	}

	c.PartyResponse = clientResponse

	if c.DataOptConfig.SqlLogs {
		c.logSql(duration)
	}

	return c
}

func (c ClientContext) logSql(duration string) {
	_fnName := "ClientParty.HitExternalApi"

	_step := GetStepInt(c.ClientRequest.ResponseId)
	_duration := duration

	type requestData struct {
		Url         string `json:"url"`
		HttpMethod  string `json:"http_method"`
		Header      string `json:"header"`
		RequestBody string `json:"request_body"`
		QueryParam  string `json:"query_param"`
		BaseAuth    string `json:"base_auth"`
	}

	type responseData struct {
		HttpCode       string `json:"http_code"`
		ResponseHeader string `json:"response_header"`
		ResponseBody   string `json:"response_body"`
	}

	type logData struct {
		RequestData  string `json:"request_data"`
		ResponseData string `json:"response_data"`
	}

	_requestData := requestData{
		Url:         c.ClientRequest.URL,
		HttpMethod:  c.ClientRequest.HttpMethod,
		Header:      fmt.Sprintf("%v", &c.ClientRequest.Header),
		RequestBody: "",
		QueryParam:  "",
		BaseAuth:    "",
	}

	if c.ClientRequest.RequestBody != nil {
		_requestData.RequestBody = *c.ClientRequest.RequestBody
	}

	if c.ClientRequest.QueryParam != nil {
		_requestData.QueryParam = fmt.Sprintf("%v", *c.ClientRequest.QueryParam)
	}

	if c.ClientRequest.Username != nil && c.ClientRequest.Password != nil {
		_requestData.BaseAuth = *c.ClientRequest.Username + ":" + *c.ClientRequest.Password
	}

	var _requestDataStr string
	_requestDataByte, err := json.Marshal(_requestData)
	if err != nil {
		_requestDataStr = fmt.Sprintf("%v", _requestData)
	} else {
		_requestDataStr = string(_requestDataByte)
	}

	var _responseData responseData
	if c.PartyResponse != nil {
		_responseData.HttpCode = strconv.Itoa(c.PartyResponse.HttpCode)
		_responseData.ResponseHeader = fmt.Sprintf("%v", c.PartyResponse.ResponseHeader)
		_responseData.ResponseBody = c.PartyResponse.ResponseBody
	}

	var _responseDataStr string
	_responseDataByte, err := json.Marshal(_responseData)
	if err != nil {
		_responseDataStr = fmt.Sprintf("%v", _responseData)
	} else {
		_responseDataStr = string(_responseDataByte)
	}

	_logData := logData{
		RequestData:  _requestDataStr,
		ResponseData: _responseDataStr,
	}

	var _logDataStr string
	_logDataByte, err := json.Marshal(_logData)
	if err != nil {
		_logDataStr = fmt.Sprintf("%v", _logData)
	} else {
		_logDataStr = string(_logDataByte)
	}

	go func() {

		data := SqlLog{
			ResponseID:   strconv.FormatInt(c.ClientRequest.ResponseId, 10),
			Step:         _step,
			Code:         "",
			Message:      "",
			FunctionName: _fnName,
			Data:         _logDataStr,
			Tracer:       "",
			Duration:     _duration,
		}

		_ = infra.GormDB.Debug().Create(&data)
	}()
}
