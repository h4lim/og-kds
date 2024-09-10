package validator

import (
	"errors"
	"strings"
	"time"

	"github.com/h4lim/og-kds/http"
)

type StringValidator interface {
	Required(errResponse http.Response) StringValidator
	MaxLength(maxLength int, errResponse http.Response) StringValidator
	IsInList(list string, errResponse http.Response) StringValidator
	IsMustSameWith(allowedValue string, errResponse http.Response) StringValidator
	IsMustNotSameWith(disallowedValue string, errResponse http.Response) StringValidator
	IsISO8601(errResponse http.Response) StringValidator
}

type StringValidatorContext struct {
	*validatorContext
	Key      string
	ValueStr string
}

func (vs *StringValidatorContext) Required(errResponse http.Response) StringValidator {
	if vs.Response.Error == nil && vs.ValueStr == "" {
		errors := errors.New(vs.Key + " is required")
		errResponse.Error = &errors

		vs.Response = errResponse
	}
	return vs
}

func (vs *StringValidatorContext) MaxLength(maxLength int, errResponse http.Response) StringValidator {
	if vs.Response.Error == nil && len(vs.ValueStr) > maxLength {
		errors := errors.New(vs.Key + " letters is exceed max letter length")
		errResponse.Error = &errors

		vs.Response = errResponse
	}
	return vs
}

func (vs *StringValidatorContext) IsInList(list string, errResponse http.Response) StringValidator {
	if vs.Response.Error == nil {
		if !strings.Contains(list, strings.ToUpper(vs.ValueStr)) {
			errors := errors.New(vs.Key + " must be in allowed list [ " + list + " ]")

			errResponse.Error = &errors
			vs.Response = errResponse
		}

	}
	return vs
}

func (vs *StringValidatorContext) IsMustSameWith(allowedValue string, errResponse http.Response) StringValidator {
	if vs.Response.Error == nil && vs.ValueStr != allowedValue {
		errors := errors.New(vs.Key + " must be the same as the allowed value")
		errResponse.Error = &errors

		vs.Response = errResponse

	}
	return vs
}

func (vs *StringValidatorContext) IsMustNotSameWith(disallowedValue string, errResponse http.Response) StringValidator {
	if vs.Response.Error == nil && vs.ValueStr == disallowedValue {
		errors := errors.New(vs.Key + " must not be the same as the disallowed value")
		errResponse.Error = &errors

		vs.Response = errResponse

	}
	return vs
}

func (vs *StringValidatorContext) IsISO8601(errResponse http.Response) StringValidator {
	if vs.Response.Error == nil {
		_, err := time.Parse(time.RFC3339, vs.ValueStr)
		if err != nil {
			errors := errors.New(vs.Key + " must be a valid ISO 8601 date")
			errResponse.Error = &errors

			vs.Response = errResponse
		}
	}
	return vs
}
