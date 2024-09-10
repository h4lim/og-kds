package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

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

type mwContext struct {
	DataOptConfig OptConfig
}

type IMw interface {
	CorsPolicy(c *gin.Context)
	DeliveryHandler(c *gin.Context)
}

func NewMw(optConfig ...OptConfig) IMw {
	var _optConfig OptConfig
	if len(optConfig) > 0 {
		_optConfig = optConfig[len(optConfig)-1]
	}

	return mwContext{
		DataOptConfig: _optConfig,
	}
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
	unixResponse := make(map[int64]int64)
	step := make(map[int64]int)
	unixResponse[responseId] = responseId
	step[responseId] = 1

	UnixTimestamp = unixResponse
	Step = step

	rawData, errGetRawData := c.GetRawData()
	duration := time.Now().UnixNano() - responseId
	ms := duration / int64(time.Millisecond)

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
			c.AbortWithStatusJSON(http.StatusInternalServerError, nil)
			return
		} else {
			zapFields = append(zapFields, zap.String("request-body", string(rawData)))
		}
		infra.ZapLog.Debug(strconv.FormatInt(responseId, 10), zapFields...)
	}

	if m.DataOptConfig.SqlLogs {

		logEntry := MwLogRequestData{
			URL:           c.Request.Method + "[" + c.Request.RequestURI + "]",
			RequestBody:   string(rawData),
			RequestHeader: c.Request.Header,
		}

		// Marshal the struct to JSON
		jsonData, err := json.Marshal(logEntry)
		if err != nil {
			fmt.Println("Error marshaling log entry to JSON:", err)
			return
		}

		jsonString := string(jsonData)

		tracer := Tracer()

		data := SqlLog{
			ResponseID:   strconv.FormatInt(responseId, 10),
			Step:         1,
			Code:         "0",
			Message:      "Success",
			FunctionName: tracer.FunctionName,
			Data:         jsonString,
			Duration:     fmt.Sprintf("%v", ms) + " ms",
			Tracer:       tracer.FileName + ":" + strconv.Itoa(tracer.Line),
		}

		_ = infra.GormDB.Debug().Create(&data)
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))
	c.Set("response-id", responseId)
	c.Next()
}
