package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/h4lim/og-kds/infra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MwLogRequestData struct {
	URL           string              `json:"url"`
	RequestBody   string              `json:"request_body"`
	RequestHeader map[string][]string `json:"request_header"`
}

type MwMqttRequestData struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

type mwContext struct {
}

type IMw interface {
	CorsPolicy(c *gin.Context)
	DeliveryHandler(c *gin.Context)
	MqttSubscribeHandler(msg mqtt.Message) int64
}

func NewMw() IMw {
	return mwContext{}
}

func (m mwContext) CorsPolicy(c *gin.Context) {

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, "+
		"Authorization, accept, Origin, Cache-Control, X-Requested-With, Access-ID,"+
		"Host, Connection, Pragma, Cache-Control, sec-ch-ua-mobile, User-Agent, sec-ch-ua, Sec-Fetch-Site, Sec-Fetch-Mode, Sec-Fetch-Dest, Referer")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

func (m mwContext) DeliveryHandler(c *gin.Context) {

	responseId := time.Now().UnixNano()

	// unixResponse := make(map[int64]int64)
	// step := make(map[int64]int)
	// requestId := make(map[int64]string)

	UnixTimestamp[responseId] = responseId
	Step[responseId] = 1

	// UnixTimestamp = unixResponse
	// Step = step

	rawData, errGetRawData := c.GetRawData()
	duration := time.Now().UnixNano() - responseId
	ms := duration / int64(time.Millisecond)

	_requestId := GetRequestIdFromRequest(rawData)
	RequestId[responseId] = _requestId
	// RequestId = requestId

	if infra.ZapLog != nil {
		zapFields := []zapcore.Field{}
		zapFields = append(zapFields, zap.Int("step", 1))

		zapFields = append(zapFields, zap.String("duration", fmt.Sprintf("%v", ms)+" ms"))
		zapFields = append(zapFields, zap.String("total-duration", fmt.Sprintf("%v", ms)+" ms"))
		zapFields = append(zapFields, zap.String("client-ip", c.ClientIP()))
		zapFields = append(zapFields, zap.String("http-method", c.Request.Method))
		zapFields = append(zapFields, zap.String("url", c.Request.RequestURI))
		zapFields = append(zapFields, zap.String("header", fmt.Sprintf("%v", c.Request.Header)))

		if errGetRawData != nil {
			zapFields = append(zapFields, zap.String("error", errGetRawData.Error()))
			infra.ZapLog.Error(strconv.FormatInt(responseId, 10), zapFields...)
			c.AbortWithStatusJSON(http.StatusInternalServerError, nil)
			return
		} else {
			zapFields = append(zapFields, zap.String("request-body", string(rawData)))
			infra.ZapLog.Debug(strconv.FormatInt(responseId, 10), zapFields...)
		}

	}

	if OptConfig.SqlLogs {

		logEntry := MwLogRequestData{
			URL:           c.Request.Method + "[" + c.Request.RequestURI + "]",
			RequestBody:   string(rawData),
			RequestHeader: c.Request.Header,
		}

		jsonString := string(jsonMarshal(logEntry))
		tracer := Tracer()

		data := sqlLog{
			ResponseID:   strconv.FormatInt(responseId, 10),
			Step:         1,
			Code:         "0",
			Message:      "Success",
			FunctionName: tracer.FunctionName,
			Data:         jsonString,
			Duration:     fmt.Sprintf("%v", ms) + " ms",
			Tracer:       tracer.FileName + ":" + strconv.Itoa(tracer.Line),
			RequestID:    _requestId,
		}

		go func() {
			_ = infra.GormDB.Debug().Create(&data)
		}()
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))
	c.Set("response-id", responseId)
	c.Next()
}

func (m mwContext) MqttSubscribeHandler(msg mqtt.Message) int64 {
	responseId := time.Now().UnixNano()

	// unixResponse := make(map[int64]int64)
	// step := make(map[int64]int)
	// requestId := make(map[int64]string)

	UnixTimestamp[responseId] = responseId
	Step[responseId] = 1

	// UnixTimestamp = unixResponse
	// Step = step

	rawData := msg.Payload()
	duration := time.Now().UnixNano() - responseId
	ms := duration / int64(time.Millisecond)

	_requestId := GetRequestIdFromRequest(rawData)
	RequestId[responseId] = _requestId
	// RequestId = requestId

	if infra.ZapLog != nil {
		zapFields := []zapcore.Field{}
		zapFields = append(zapFields, zap.Int("step", 1))

		zapFields = append(zapFields, zap.String("duration", fmt.Sprintf("%v", ms)+" ms"))
		zapFields = append(zapFields, zap.String("total-duration", fmt.Sprintf("%v", ms)+" ms"))
		zapFields = append(zapFields, zap.String("mqtt-topic", msg.Topic()))

		zapFields = append(zapFields, zap.String("mqtt-payload", string(rawData)))
		infra.ZapLog.Debug(strconv.FormatInt(responseId, 10), zapFields...)

	}

	if OptConfig.SqlLogs {

		logEntry := MwMqttRequestData{
			Topic:   msg.Topic(),
			Payload: string(rawData),
		}

		jsonString := string(jsonMarshal(logEntry))
		tracer := Tracer()

		data := sqlLog{
			ResponseID:   strconv.FormatInt(responseId, 10),
			Step:         1,
			Code:         "0",
			Message:      "Success",
			FunctionName: tracer.FunctionName,
			Data:         jsonString,
			Duration:     fmt.Sprintf("%v", ms) + " ms",
			Tracer:       tracer.FileName + ":" + strconv.Itoa(tracer.Line),
			RequestID:    _requestId,
		}

		go func() {
			_ = infra.GormDB.Debug().Create(&data)
		}()
	}

	return responseId
}

// func (m mwContext) SetInitialTracerData(_requestId string) {

// 	unixResponse := make(map[int64]int64)
// 	step := make(map[int64]int)
// 	requestId := make(map[int64]string)

// 	responseId := time.Now().UnixNano()
// 	unixResponse[responseId] = responseId
// 	step[responseId] = 1
// 	requestId[responseId] = _requestId

// 	UnixTimestamp = unixResponse
// 	Step = step
// 	RequestId = requestId

// 	duration := time.Now().UnixNano() - responseId
// 	ms := duration / int64(time.Millisecond)
// }
