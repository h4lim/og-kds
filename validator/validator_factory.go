package validator

import (
	"github.com/h4lim/og-kds/http"
)

type Validator interface {
	Validate(key string, value interface{}) Validator
}

type validatorContext struct {
	OptionalData http.OptSetR
	Error        *error
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
	return vc.Error != nil
}

func (vc *validatorContext) GetError() *error {
	return vc.Error
}
