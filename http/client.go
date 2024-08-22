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
}

type IClient interface {
	Hit() ClientContext
	MustHttpOk200() ClientContext
	UnmarshalJson(jsonData *interface{}) ClientContext
	GetPartyResponse() (*party.Response, error)
}

func NewClient(ctx ClientRequest) IClient {
	return ClientContext{
		ClientRequest: ctx,
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
			c.Error = errors.New("http must 200")
		}
	}

	return c
}

// UnmarshalJson implements IClient.
func (c ClientContext) UnmarshalJson(jsonData *interface{}) ClientContext {
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

	zapFields = append(zapFields, zap.String("step", GetNextStep(c.ClientRequest.ResponseId)))
	zapFields = append(zapFields, zap.String("duration", GetDuration(c.ClientRequest.ResponseId)+" ms"))

	duration := time.Now().UnixNano() - c.ClientRequest.ResponseId
	ms := float64(duration) / float64(time.Millisecond)
	zapFields = append(zapFields, zap.String("total-duration", fmt.Sprintf("%v", ms)+" ms"))

	if c.ClientRequest.AdditionalTracer != nil {
		zapFields = append(zapFields, zap.String("additional-tracer", strings.Join(*c.ClientRequest.AdditionalTracer, " ")))
	}

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

	zapFields = append(zapFields, zap.Int("http-code",
		clientResponse.HttpCode))

	zapFields = append(zapFields, zap.String("client-response",
		clientResponse.ResponseBody))

	if infra.ZapLog != nil {
		infra.ZapLog.Debug(strconv.FormatInt(c.ClientRequest.ResponseId, 10), zapFields...)
	}

	c.PartyResponse = clientResponse

	return c
}
