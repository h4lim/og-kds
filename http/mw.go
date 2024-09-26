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
	HttpMethod    string              `json:"http_method"`
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

	UnixTimestamp[responseId] = responseId
	Step[responseId] = 1

	rawData, errGetRawData := c.GetRawData()
	duration := time.Now().UnixNano() - responseId
	ms := duration / int64(time.Millisecond)

	_requestId := GetRequestIdFromRequest(rawData)
	RequestId[responseId] = _requestId

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
			infra.ZapLog.Warn(strconv.FormatInt(responseId, 10), zapFields...)
			c.AbortWithStatusJSON(http.StatusInternalServerError, nil)
			return
		} else {
			zapFields = append(zapFields, zap.String("request-body", string(rawData)))
			infra.ZapLog.Debug(strconv.FormatInt(responseId, 10), zapFields...)
		}

	}

	if OptConfig.SqlLogs {

		logEntry := MwLogRequestData{
			HttpMethod:    c.Request.Method,
			URL:           c.Request.RequestURI,
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
			FunctionName: getFunctionName(tracer.FunctionName),
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

	UnixTimestamp[responseId] = responseId
	Step[responseId] = 1

	rawData := msg.Payload()
	duration := time.Now().UnixNano() - responseId
	ms := duration / int64(time.Millisecond)

	_requestId := GetRequestIdFromRequest(rawData)
	RequestId[responseId] = _requestId

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
			FunctionName: getFunctionName(tracer.FunctionName),
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
