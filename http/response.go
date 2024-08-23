package http

import (
	"fmt"
	"github.com/h4lim/og-kds/infra"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RequestBuildGin struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	ResponseID int64  `json:"response_id"`
}

type RequestBuildGinWithData struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	ResponseID int64  `json:"response_id"`
}

type RequestBuildGinSnap struct {
	ResponseCode    int    `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	ReferenceNo     int64  `json:"referenceNo"`
}

type RequestBuildGinSnapWithData struct {
	ResponseCode    int    `json:"responseCode"`
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
	Code             int
	AdditionalCode   int
	Message          string
	Data             any
	Error            *error
	Tracer           TracerModel
	ResponseID       int64
	Language         string
	AdditionalTracer []string
}

func InitResponse(responseID int64, language string) Response {
	return Response{
		ResponseID: responseID,
		Language:   language,
	}
}

func (r Response) SetAdditionalTracer(additionalTracer string) Response {
	r.AdditionalTracer = append(r.AdditionalTracer, additionalTracer)
	return r
}

func (r Response) SetAll(newR Response) Response {

	if newR.Error != nil {
		if newR.HttpCode == 0 {
			newR.HttpCode = 400
		}

		if newR.Code == 0 {
			newR.Code = 99
		}
	} else {
		if newR.HttpCode == 0 {
			newR.HttpCode = 200
		}
	}

	r.HttpCode = newR.HttpCode
	r.Code = newR.Code
	r.AdditionalCode = newR.AdditionalCode
	r.Message = newR.Message
	r.Data = newR.Data
	r.Error = newR.Error
	r.Tracer = newR.Tracer

	r.getMessage()
	r.debug(true)

	return r
}

func (r Response) SetCode(newCode int) Response {
	previousCode := r.Code
	r.AdditionalCode = 0
	r.Code = newCode

	r.getMessage()

	if infra.ZapLog != nil {
		zapFields := []zapcore.Field{}
		zapFields = append(zapFields, zap.String("code-info", "Remapping from "+strconv.Itoa(previousCode)+" to "+strconv.Itoa(r.Code)))
		zapFields = append(zapFields, zap.String("code", strconv.Itoa(r.Code)))
		zapFields = append(zapFields, zap.String("message ", r.Message))
		infra.ZapLog.Debug(strconv.FormatInt(r.ResponseID, 10), zapFields...)
	}

	r.debug(false)

	return r
}

func (r Response) SetError(newError *error) Response {

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

		infra.ZapLog.Debug(strconv.FormatInt(r.ResponseID, 10), zapFields...)
	}

	r.debug(false)

	return r
}

func (r Response) BuildGinResponse() (int, any) {

	return r.HttpCode, RequestBuildGin{
		Code:       r.Code,
		Message:    r.Message,
		ResponseID: r.ResponseID,
	}
}

func (r Response) BuildGinResponseWithData(data any) (int, any) {
	r.Data = data

	return r.HttpCode, RequestBuildGinWithData{
		Code:       r.Code,
		Message:    r.Message,
		ResponseID: r.ResponseID,
		Data:       r.Data,
	}
}

func (r Response) BuildGinResponseSnap() (int, any) {

	return r.HttpCode, RequestBuildGinSnap{
		ResponseCode:    r.Code,
		ResponseMessage: r.Message,
		ReferenceNo:     r.ResponseID,
	}
}

func (r Response) BuildGinResponseSnapWithData(data any) (int, any) {
	r.Data = data

	return r.HttpCode, RequestBuildGinSnapWithData{
		ResponseCode:    r.Code,
		ResponseMessage: r.Message,
		ReferenceNo:     r.ResponseID,
		Data:            r.Data,
	}
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
		zapFields = append(zapFields, zap.Int("code", r.Code))
		zapFields = append(zapFields, zap.String("message ", r.Message))

		if r.Error != nil {
			zapFields = append(zapFields, zap.String("error", fmt.Sprintf("%v", *r.Error)))
		}

		zapFields = append(zapFields, zap.String("data", fmt.Sprintf("%v", r.Data)))
		zapFields = append(zapFields, zap.String("filename", r.Tracer.FileName))
		zapFields = append(zapFields, zap.String("function-name", r.Tracer.FunctionName))
		zapFields = append(zapFields, zap.Int("line", r.Tracer.Line))
		zapFields = append(zapFields, zap.String("trace", r.Tracer.FileName+":"+strconv.Itoa(r.Tracer.Line)))

		infra.ZapLog.Debug(strconv.FormatInt(r.ResponseID, 10), zapFields...)
	}

}

func (r *Response) getMessage() {
	strCode := strconv.Itoa(r.Code)

	if r.AdditionalCode != 0 {
		strCode = strCode + "_" + strconv.Itoa(r.AdditionalCode)
	}

	switch {
	case strings.ToUpper(r.Language) == "ID":
		r.Message = infra.MessageID[strCode]
	case strings.ToUpper(r.Language) == "EN":
		r.Message = infra.MessageEN[strCode]
	default:
		r.Message = infra.MessageEN["EN"]
	}

	if r.Message == "" {
		r.Message = "unknown message"
	}
}
