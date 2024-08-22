package http

import (
	"fmt"
	party "github.com/h4lim/client-party"
	"github.com/h4lim/og-kds/infra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strconv"
	"strings"
	"time"
)

type ClientContext struct {
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

type IClient interface {
	Hit() (*party.Response, error)
}

func NewClient(ctx ClientContext) IClient {
	return ctx
}

func (c ClientContext) Hit() (*party.Response, error) {

	clientParty := party.NewClientParty(c.HttpMethod, c.URL).
		SetHeader(c.Header["Content-Type"], c.Header)

	var zapFields []zapcore.Field

	zapFields = append(zapFields, zap.String("step", GetNextStep(c.ResponseId)))
	zapFields = append(zapFields, zap.String("duration", GetDuration(c.ResponseId)+" ms"))

	duration := time.Now().UnixNano() - c.ResponseId
	ms := float64(duration) / float64(time.Millisecond)
	zapFields = append(zapFields, zap.String("total-duration", fmt.Sprintf("%v", ms)+" ms"))

	if c.AdditionalTracer != nil {
		zapFields = append(zapFields, zap.String("additional-tracer", strings.Join(*c.AdditionalTracer, " ")))
	}

	if c.RequestBody != nil {
		clientParty.SetRequestBodyStr(*c.RequestBody)
		zapFields = append(zapFields, zap.String("request-body", *c.RequestBody))
	}

	if c.QueryParam != nil {
		clientParty.SetQueryParam(*c.QueryParam)
		zapFields = append(zapFields, zap.String("query-param",
			fmt.Sprintf("%v", *c.QueryParam)))
	}

	if c.Username != nil && c.Password != nil {
		clientParty.SetBaseAuth(*c.Username, *c.Password)
	}

	clientResponse, err := clientParty.HitClient()
	if err != nil {
		zapFields = append(zapFields, zap.String("query-param",
			fmt.Sprintf("%v", *err)))
		return nil, *err
	}

	zapFields = append(zapFields, zap.Int("http-code",
		clientResponse.HttpCode))

	zapFields = append(zapFields, zap.String("client-response",
		clientResponse.ResponseBody))

	if infra.ZapLog != nil {
		infra.ZapLog.Debug(strconv.FormatInt(c.ResponseId, 10), zapFields...)
	}

	return clientResponse, nil
}
