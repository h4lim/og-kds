package validator

import (
	"errors"
	"fmt"

	"github.com/h4lim/og-kds/http"
)

type IntValidator interface {
	Required(errResponse http.Response) IntValidator
	GreaterThan(threshold int, errResponse http.Response) IntValidator
	LessThan(threshold int, errResponse http.Response) IntValidator
}

type intValidatorContext struct {
	*validatorContext
	Key      string
	ValueInt int64
}

func (iv *intValidatorContext) Required(errResponse http.Response) IntValidator {
	if iv.Response.Error == nil && iv.ValueInt == 0 {
		errors := errors.New(iv.Key + " is required")
		errResponse.Error = &errors

		iv.Response = errResponse
	}
	return iv
}

func (iv *intValidatorContext) GreaterThan(threshold int, errResponse http.Response) IntValidator {
	if iv.Response.Error == nil && iv.ValueInt < int64(threshold) {
		errors := errors.New(iv.Key + " is must greater than " + fmt.Sprintf("%d", threshold))
		errResponse.Error = &errors

		iv.Response = errResponse
	}
	return iv
}

func (iv *intValidatorContext) LessThan(threshold int, errResponse http.Response) IntValidator {
	if iv.Response.Error == nil && iv.ValueInt > int64(threshold) {
		errors := errors.New(iv.Key + " is must less than " + fmt.Sprintf("%d", threshold))
		errResponse.Error = &errors

		iv.Response = errResponse
	}
	return iv
}
