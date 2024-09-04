package http

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/h4lim/og-kds/infra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"strconv"
	"time"
)

type mwContext struct {
}

type IMw interface {
	CorsPolicy(c *gin.Context)
	DeliveryHandler(c *gin.Context)
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
	unixResponse := make(map[int64]int64)
	step := make(map[int64]int)
	unixResponse[responseId] = responseId
	step[responseId] = 1

	UnixTimestamp = unixResponse
	Step = step

	if infra.ZapLog != nil {
		zapFields := []zapcore.Field{}
		zapFields = append(zapFields, zap.Int("step", 1))

		duration := time.Now().UnixNano() - responseId
		ms := duration / int64(time.Millisecond)
		zapFields = append(zapFields, zap.String("duration", fmt.Sprintf("%v", ms)+" ms"))
		zapFields = append(zapFields, zap.String("total-duration", fmt.Sprintf("%v", ms)+" ms"))
		zapFields = append(zapFields, zap.String("client-ip", c.ClientIP()))
		zapFields = append(zapFields, zap.String("http-method", c.Request.Method))
		zapFields = append(zapFields, zap.String("url", c.Request.RequestURI))
		zapFields = append(zapFields, zap.String("header", fmt.Sprintf("%v", c.Request.Header)))

		rawData, err := c.GetRawData()
		if err != nil {
			zapFields = append(zapFields, zap.String("error", err.Error()))
			c.AbortWithStatusJSON(http.StatusInternalServerError, nil)
			return
		} else {
			zapFields = append(zapFields, zap.String("request-body", string(rawData)))
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawData))
		infra.ZapLog.Debug(strconv.FormatInt(responseId, 10), zapFields...)
	}

	c.Set("response-id", responseId)
	c.Next()
}
