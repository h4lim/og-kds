package validator

import (
	"errors"
	"fmt"

	"github.com/h4lim/og-kds/http"
)

type FloatValidator interface {
	Required(errResponse http.OptSetR) FloatValidator
	GreaterThan(threshold float64, errResponse http.OptSetR) FloatValidator
	LessThan(threshold float64, errResponse http.OptSetR) FloatValidator
	MustNot(disallowedValue float64, errResponse http.OptSetR) FloatValidator
}

type floatValidatorContext struct {
	*validatorContext
	Key      string
	ValueInt float64
}

func (fv *floatValidatorContext) Required(errResponse http.OptSetR) FloatValidator {
	if fv.Error == nil && fv.ValueInt == 0 {
		errors := errors.New(fv.Key + " is required")

		fv.OptionalData = errResponse
		fv.Error = &errors
	}
	return fv
}

func (fv *floatValidatorContext) GreaterThan(threshold float64, errResponse http.OptSetR) FloatValidator {
	if fv.Error == nil && fv.ValueInt < threshold {
		errors := errors.New(fv.Key + " is must greater than " + fmt.Sprintf("%f", threshold))

		fv.OptionalData = errResponse
		fv.Error = &errors
	}
	return fv
}

func (fv *floatValidatorContext) LessThan(threshold float64, errResponse http.OptSetR) FloatValidator {
	if fv.Error == nil && fv.ValueInt > threshold {
		errors := errors.New(fv.Key + " is must less than " + fmt.Sprintf("%f", threshold))

		fv.OptionalData = errResponse
		fv.Error = &errors
	}
	return fv
}

func (fv *floatValidatorContext) MustNot(disallowedValue float64, errResponse http.OptSetR) FloatValidator {
	if fv.Error == nil && fv.ValueInt == disallowedValue {
		errors := errors.New(fv.Key + " is must not be " + fmt.Sprintf("%f", disallowedValue))

		fv.OptionalData = errResponse
		fv.Error = &errors
	}
	return fv
}
