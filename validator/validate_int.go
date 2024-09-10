package validator

import (
	"errors"
	"fmt"

	"github.com/h4lim/og-kds/http"
)

type IntValidator interface {
	Required(errResponse http.OptSetR) IntValidator
	GreaterThan(threshold int, errResponse http.OptSetR) IntValidator
	LessThan(threshold int, errResponse http.OptSetR) IntValidator
}

type intValidatorContext struct {
	*validatorContext
	Key      string
	ValueInt int64
}

func (iv *intValidatorContext) Required(errResponse http.OptSetR) IntValidator {
	if iv.Error == nil && iv.ValueInt == 0 {
		errors := errors.New(iv.Key + " is required")

		iv.OptionalData = errResponse
		iv.Error = &errors
	}
	return iv
}

func (iv *intValidatorContext) GreaterThan(threshold int, errResponse http.OptSetR) IntValidator {
	if iv.Error == nil && iv.ValueInt < int64(threshold) {
		errors := errors.New(iv.Key + " is must greater than " + fmt.Sprintf("%d", threshold))

		iv.OptionalData = errResponse
		iv.Error = &errors
	}
	return iv
}

func (iv *intValidatorContext) LessThan(threshold int, errResponse http.OptSetR) IntValidator {
	if iv.Error == nil && iv.ValueInt > int64(threshold) {
		errors := errors.New(iv.Key + " is must less than " + fmt.Sprintf("%d", threshold))

		iv.OptionalData = errResponse
		iv.Error = &errors
	}
	return iv
}
