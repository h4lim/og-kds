package http

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/h4lim/og-kds/infra"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RequestBuildGin struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	ResponseID int64  `json:"response_id"`
}

type RequestBuildGinWithData struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	ResponseID int64  `json:"response_id"`
}

type RequestBuildGinSnap struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	ReferenceNo     int64  `json:"referenceNo"`
}

type RequestBuildGinSnapWithData struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	ReferenceNo     int64  `json:"referenceNo"`
	Data            any    `json:"data"`
}

type ResponseBuildGin struct {
	HttpCode int
	Obj      any
}

type TracerModel struct {
	FunctionName  string `json:"function_name"`
	FileName      string `json:"file_name"`
	Line          int    `json:"line"`
	Step          int    `json:"step"`
	Duration      string `json:"duration"`
	TotalDuration string `json:"total_duration"`
}

type Response struct {
	HttpCode         int
	Code             string
	Message          string
	Data             any
	Error            *error
	Tracer           TracerModel
	ResponseID       int64
	Language         string
	AdditionalTracer []string
}

type OptSetR struct {
	HttpCode int
	Code     string
	Message  string
	Data     any
}

func InitResponse(responseID int64, language string) Response {
	return Response{
		ResponseID: responseID,
		Language:   language,
	}
}

func (r *Response) SetSuccessR(Tracer TracerModel, optData ...OptSetR) Response {
	r.Tracer = Tracer

	var _optData OptSetR
	if len(optData) > 0 {
		_optData = optData[len(optData)-1]
	}

	if _optData.HttpCode == 0 {
		_optData.HttpCode = 200
	}

	if _optData.Code == "" {
		_optData.Code = "0"
	}

	r.HttpCode = _optData.HttpCode
	r.Code = _optData.Code
	r.Message = _optData.Message
	r.Data = _optData.Data

	if r.Message == "" {
		r.getMessage()
	}

	r.debug(true)

	if OptConfig.SqlLogs {
		r.logSql()
	}

	return *r
}

func (r *Response) SetErrorR(Error *error, Tracer TracerModel, optData ...OptSetR) Response {
	r.Error = Error
	r.Tracer = Tracer

	var _optData OptSetR
	if len(optData) > 0 {
		_optData = optData[len(optData)-1]
	}

	if _optData.HttpCode == 0 {
		_optData.HttpCode = 400
	}

	if _optData.Code == "" {
		_optData.Code = "99"
	}

	r.HttpCode = _optData.HttpCode
	r.Code = _optData.Code
	r.Message = _optData.Message
	r.Data = _optData.Data

	if r.Message == "" {
		r.getMessage()
	}

	r.debug(true)

	if OptConfig.SqlLogs {
		r.logSql()
	}

	return *r
}

func (r *Response) SetAdditionalTracer(additionalTracer string) Response {
	r.AdditionalTracer = append(r.AdditionalTracer, additionalTracer)
	return *r
}

func (r *Response) SetAll(newR Response) Response {

	if newR.Error != nil {
		if newR.HttpCode == 0 {
			newR.HttpCode = 400
		}

		if newR.Code == "" {
			newR.Code = "99"
		}
	} else {
		if newR.HttpCode == 0 {
			newR.HttpCode = 200
			newR.Code = "0"
		}
	}

	r.HttpCode = newR.HttpCode
	r.Code = newR.Code
	r.Message = newR.Message
	r.Data = newR.Data
	r.Error = newR.Error
	r.Tracer = newR.Tracer

	if r.Message == "" {
		r.getMessage()
	}

	r.debug(true)

	if OptConfig.SqlLogs {
		r.logSql()
	}

	return *r
}

func (r Response) SetCode(newCode string) Response {
	previousCode := r.Code
	r.Code = newCode

	r.getMessage()

	if infra.ZapLog != nil {
		zapFields := []zapcore.Field{}
		zapFields = append(zapFields, zap.String("code-info", "Remapping from "+previousCode+" to "+r.Code))
		zapFields = append(zapFields, zap.String("code", r.Code))
		zapFields = append(zapFields, zap.String("message ", r.Message))
		infra.ZapLog.Debug(strconv.FormatInt(r.ResponseID, 10), zapFields...)
	}

	r.debug(false)

	return r
}

func (r *Response) SetError(newError *error) Response {

	previousError := r.Error
	r.Error = newError

	if infra.ZapLog != nil {

		zapFields := []zapcore.Field{}
		if previousError != nil {
			zapFields = append(zapFields, zap.String("error-info", "Remapping from "+
				fmt.Sprintf("%v", *previousError)+" to "+fmt.Sprintf("%v", *r.Error)))
		} else {
			zapFields = append(zapFields, zap.String("error-info", "New Error "+fmt.Sprintf("%v", *r.Error)))
		}

		infra.ZapLog.Warn(strconv.FormatInt(r.ResponseID, 10), zapFields...)
	}

	r.debug(false)

	return *r
}

func (r Response) BuildGinResponse() (int, any) {

	r.Tracer.FunctionName = "api.finalResponse"
	r.debug(true)
	if OptConfig.SqlLogs {
		r.logSql()
	}

	delete(UnixTimestamp, r.ResponseID)
	delete(Step, r.ResponseID)
	delete(RequestId, r.ResponseID)

	return r.HttpCode, RequestBuildGin{
		Code:       r.Code,
		Message:    r.Message,
		ResponseID: r.ResponseID,
	}
}

func (r Response) BuildGinResponseWithData(data any) (int, any) {

	r.Data = data
	r.Tracer.FunctionName = "api.finalResponse"
	r.debug(true)
	if OptConfig.SqlLogs {
		r.logSql()
	}

	delete(UnixTimestamp, r.ResponseID)
	delete(Step, r.ResponseID)
	delete(RequestId, r.ResponseID)

	return r.HttpCode, RequestBuildGinWithData{
		Code:       r.Code,
		Message:    r.Message,
		ResponseID: r.ResponseID,
		Data:       r.Data,
	}
}

func (r Response) BuildGinResponseSnap() (int, any) {

	r.Tracer.FunctionName = "api.finalResponse"
	r.debug(true)
	if OptConfig.SqlLogs {
		r.logSql()
	}

	delete(UnixTimestamp, r.ResponseID)
	delete(Step, r.ResponseID)
	delete(RequestId, r.ResponseID)

	return r.HttpCode, RequestBuildGinSnap{
		ResponseCode:    r.Code,
		ResponseMessage: r.Message,
		ReferenceNo:     r.ResponseID,
	}
}

func (r Response) BuildGinResponseSnapWithData(data any) (int, any) {

	r.Data = data
	r.Tracer.FunctionName = "api.finalResponse"
	r.debug(true)
	if OptConfig.SqlLogs {
		r.logSql()
	}

	delete(UnixTimestamp, r.ResponseID)
	delete(Step, r.ResponseID)
	delete(RequestId, r.ResponseID)

	return r.HttpCode, RequestBuildGinSnapWithData{
		ResponseCode:    r.Code,
		ResponseMessage: r.Message,
		ReferenceNo:     r.ResponseID,
		Data:            r.Data,
	}
}

func (r Response) BuildVoidResponse() {

	r.Tracer.FunctionName = "finalResponse"
	r.debug(true)
	if OptConfig.SqlLogs {
		r.logSql()
	}

	delete(UnixTimestamp, r.ResponseID)
	delete(Step, r.ResponseID)
	delete(RequestId, r.ResponseID)
}

func (r *Response) IsError() bool {
	return r.Error != nil
}

func (r *Response) debug(nextStep bool) {

	if infra.ZapLog != nil {

		duration := time.Now().UnixNano() - r.ResponseID
		ms := float64(duration) / float64(time.Millisecond)
		zapFields := []zapcore.Field{}

		if nextStep {
			zapFields = append(zapFields, zap.String("step", GetNextStep(r.ResponseID)))
		} else {
			zapFields = append(zapFields, zap.String("step", GetStep(r.ResponseID)))
		}

		zapFields = append(zapFields, zap.String("duration", GetDuration(r.ResponseID)+" ms"))
		zapFields = append(zapFields, zap.String("total-duration", fmt.Sprintf("%v", ms)+" ms"))
		zapFields = append(zapFields, zap.String("additional-tracer", strings.Join(r.AdditionalTracer, " ")))
		zapFields = append(zapFields, zap.Int("http-code", r.HttpCode))
		zapFields = append(zapFields, zap.String("code", r.Code))
		zapFields = append(zapFields, zap.String("message ", r.Message))
		zapFields = append(zapFields, zap.String("data", fmt.Sprintf("%v", r.Data)))
		zapFields = append(zapFields, zap.String("filename", r.Tracer.FileName))
		zapFields = append(zapFields, zap.String("function-name", r.Tracer.FunctionName))
		zapFields = append(zapFields, zap.Int("line", r.Tracer.Line))
		zapFields = append(zapFields, zap.String("trace", r.Tracer.FileName+":"+strconv.Itoa(r.Tracer.Line)))

		if r.Error != nil {
			zapFields = append(zapFields, zap.String("error", fmt.Sprintf("%v", *r.Error)))
			infra.ZapLog.Warn(strconv.FormatInt(r.ResponseID, 10), zapFields...)
		} else {
			infra.ZapLog.Debug(strconv.FormatInt(r.ResponseID, 10), zapFields...)
		}
	}

}

func (r *Response) getMessage() {

	switch {
	case strings.ToUpper(r.Language) == "ID":
		r.Message = infra.MessageID[r.Code]
	case strings.ToUpper(r.Language) == "EN":
		r.Message = infra.MessageEN[r.Code]
	default:
		r.Message = infra.MessageEN["EN"]
	}

	if r.Message == "" {
		r.Message = "unknown message"
	}

}

func (r *Response) logSql() {
	_fnName := getFunctionName(r.Tracer.FunctionName)

	if r.Message == "" {
		r.getMessage()
	}

	_requestId := GetRequestId(r.ResponseID)
	_step := GetStepInt(r.ResponseID)
	_duration := GetDuration(r.ResponseID) + " ms"

	var _data string
	jsonData, err := json.Marshal(r.Data)
	if err != nil {
		_data = fmt.Sprintf("%v", r.Data)
	} else {
		_data = string(jsonData)
	}

	data := sqlLog{
		ResponseID:   strconv.FormatInt(r.ResponseID, 10),
		Step:         _step,
		Code:         r.Code,
		Message:      r.Message,
		FunctionName: _fnName,
		Data:         _data,
		Tracer:       r.Tracer.FileName + ":" + strconv.Itoa(r.Tracer.Line),
		Duration:     _duration,
		RequestID:    _requestId,
	}

	go func() {
		_ = infra.GormDB.Debug().Create(&data)
	}()
}
