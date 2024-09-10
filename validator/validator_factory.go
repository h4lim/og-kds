package validator

import (
	"github.com/h4lim/og-kds/http"
)

type Validator interface {
	Validate(key string, value interface{}) Validator
}

type FloatValidator interface {
	Required(errResponse http.Response) FloatValidator
	GreaterThan(threshold float64, errResponse http.Response) FloatValidator
	LessThan(threshold float64, errResponse http.Response) FloatValidator
	MustNot(disallowedValue float64, errResponse http.Response) FloatValidator
}

type validatorContext struct {
	Response http.Response
}

func NewValidator() *validatorContext {
	return &validatorContext{}
}

func (vc *validatorContext) Validate(key string, value interface{}) Validator {
	switch v := value.(type) {
	case string:
		return &StringValidatorContext{validatorContext: vc, Key: key, ValueStr: v}
	case int64:
		return &intValidatorContext{validatorContext: vc, Key: key, ValueInt: v}
	case int:
		return &intValidatorContext{validatorContext: vc, Key: key, ValueInt: int64(v)}
	case float64:
		return &floatValidatorContext{validatorContext: vc, Key: key, ValueInt: v}
	default:
		return nil
	}
}

func (vc *validatorContext) IsError() bool {
	return vc.Response.Error != nil
}

func (vc *validatorContext) GetResponse() http.Response {
	return vc.Response
}
