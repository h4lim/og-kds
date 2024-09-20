package http

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	UnixTimestamp map[int64]int64
	Step          map[int64]int
	RequestId     map[int64]string
	OptConfig     OptConfigModel
)

type PackageInformationModel struct {
	FunctionName string
	FileName     string
	Line         int
}

type RequestHeaderSnap struct {
	ContentType           string
	AcceptLanguage        string
	AuthorizationCustomer string
	XTimestamp            string
	XSignature            string
	XPartnerId            string
	XEternalId            string
	ChannelId             string
}

func InitHttp(config OptConfigModel) {
	RequestId = make(map[int64]string)
	UnixTimestamp = make(map[int64]int64)
	Step = make(map[int64]int)

	setOptionalConfig(config)
}

func GetHeaderSnapTransaction(c *gin.Context) RequestHeaderSnap {
	var headerSnap RequestHeaderSnap

	headerSnap.ContentType = c.GetHeader("Content-Type")
	headerSnap.AuthorizationCustomer = c.GetHeader("Authorization-Customer")
	headerSnap.XTimestamp = c.GetHeader("X-Timestamp")
	headerSnap.XSignature = c.GetHeader("X-Signature")
	headerSnap.XPartnerId = c.GetHeader("X-Partner-Id")
	headerSnap.XEternalId = c.GetHeader("X-Eternal-Id")
	headerSnap.ChannelId = c.GetHeader("Channel-Id")
	headerSnap.AcceptLanguage = "EN"

	return headerSnap
}

func GetRandomAlphaNumeric(length int) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

func GenerateSession() string {
	u := uuid.New()
	re := regexp.MustCompile("[^a-zA-Z0-9]+")

	cleanUUID := re.ReplaceAllString(u.String(), "")
	upperUUID := strings.ToUpper(cleanUUID)

	session := fmt.Sprintf("%s%d", upperUUID, time.Now().UnixNano()/int64(time.Millisecond))
	return session
}

func GetDuration(responseId int64) string {

	newUnixNano := time.Now().UnixNano()
	duration := UnixTimestamp[responseId]
	elapsed := newUnixNano - duration
	UnixTimestamp[responseId] = newUnixNano
	ms := float64(elapsed) / float64(time.Millisecond)

	return fmt.Sprintf("%v", ms)
}

func GetRequestId(responseId int64) string {
	return RequestId[responseId]
}

func GetStepInt(responseId int64) int {
	return Step[responseId]
}

func GetStep(responseId int64) string {
	return strconv.Itoa(Step[responseId])
}

func GetNextStep(responseId int64) string {
	step := Step[responseId]
	Step[responseId] = step + 1
	return strconv.Itoa(Step[responseId])
}

func Tracer() TracerModel {

	var model TracerModel
	pc, fileName, line, ok := runtime.Caller(1)
	if !ok {
		return model
	}

	model.FileName = fileName
	model.Line = line

	callerFunction := runtime.FuncForPC(pc)
	if callerFunction != nil {
		model.FunctionName = callerFunction.Name()
	}

	return model
}

func GetResponseIdAndLanguage(c *gin.Context) (int64, string) {

	language := getLanguage(c)
	responseId, exist := c.Get("response-id")
	if !exist {
		return time.Now().UnixNano(), language
	}

	int64Value, ok := responseId.(int64)
	if !ok {
		return time.Now().UnixNano(), language
	}

	return int64Value, language
}

func getLanguage(c *gin.Context) string {

	language, exist := c.Get("Accept-Language")
	if !exist {
		return "EN"
	}

	strLanguage, ok := language.(string)
	if !ok {
		return "EN"
	}

	return strLanguage
}

func GetRequestIdFromRequest(rawBody []byte) string {
	var regex *regexp.Regexp
	var requestId string

	if OptConfig.RequestIdAlias != "" {
		regex = regexp.MustCompile(`"(` + OptConfig.RequestIdAlias + `)":\s*"(.*?)"`)
		match := regex.FindSubmatch(rawBody)

		if len(match) > 2 {
			requestId = string(match[2])
		}
	} else {
		regex = regexp.MustCompile(`"` + "request_id" + `":\s*"(.*?)"`)
		match := regex.FindSubmatch(rawBody)

		if len(match) > 1 {
			requestId = string(match[1])
		}
	}

	return requestId
}

func jsonMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func getFunctionName(fnName string) string {
	parts := strings.Split(fnName, "/")
	if len(parts) > 0 {
		fnName = parts[len(parts)-1]
	}

	return fnName
}
