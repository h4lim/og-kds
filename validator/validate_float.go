package validator

import (
	"errors"
	"fmt"

	"github.com/h4lim/og-kds/http"
)

type floatValidatorContext struct {
	*validatorContext
	Key      string
	ValueInt float64
}

func (fv *floatValidatorContext) Required(errResponse http.Response) FloatValidator {
	if fv.Response.Error == nil && fv.ValueInt == 0 {
		errors := errors.New(fv.Key + " is required")
		errResponse.Error = &errors

		fv.Response = errResponse
	}
	return fv
}

func (fv *floatValidatorContext) GreaterThan(threshold float64, errResponse http.Response) FloatValidator {
	if fv.Response.Error == nil && fv.ValueInt < threshold {
		errors := errors.New(fv.Key + " is must greater than " + fmt.Sprintf("%f", threshold))
		errResponse.Error = &errors

		fv.Response = errResponse
	}
	return fv
}

func (fv *floatValidatorContext) LessThan(threshold float64, errResponse http.Response) FloatValidator {
	if fv.Response.Error == nil && fv.ValueInt > threshold {
		errors := errors.New(fv.Key + " is must less than " + fmt.Sprintf("%f", threshold))
		errResponse.Error = &errors

		fv.Response = errResponse
	}
	return fv
}

func (fv *floatValidatorContext) MustNot(disallowedValue float64, errResponse http.Response) FloatValidator {
	if fv.Response.Error == nil && fv.ValueInt == disallowedValue {
		errors := errors.New(fv.Key + " is must not be " + fmt.Sprintf("%f", disallowedValue))
		errResponse.Error = &errors

		fv.Response = errResponse
	}
	return fv
}
