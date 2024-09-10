package validator

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/h4lim/og-kds/http"
)

type StringValidator interface {
	Required(errResponse http.OptSetR) StringValidator
	MaxLength(maxLength int, errResponse http.OptSetR) StringValidator
	IsInList(list string, errResponse http.OptSetR) StringValidator
	IsMustSameWith(allowedValue string, errResponse http.OptSetR) StringValidator
	IsMustNotSameWith(disallowedValue string, errResponse http.OptSetR) StringValidator
	IsAlphaNumeric(errResponse http.OptSetR) StringValidator
}

type StringValidatorContext struct {
	*validatorContext
	Key      string
	ValueStr string
}

func (vs *StringValidatorContext) Required(errResponse http.OptSetR) StringValidator {
	if vs.Error == nil && vs.ValueStr == "" {
		errors := errors.New(vs.Key + " is required")

		vs.OptionalData = errResponse
		vs.Error = &errors
	}
	return vs
}

func (vs *StringValidatorContext) MaxLength(maxLength int, errResponse http.OptSetR) StringValidator {
	if vs.Error == nil && len(vs.ValueStr) > maxLength {
		errors := errors.New(vs.Key + " letters is exceed max letter length")

		vs.OptionalData = errResponse
		vs.Error = &errors
	}
	return vs
}

func (vs *StringValidatorContext) IsInList(list string, errResponse http.OptSetR) StringValidator {
	if vs.Error == nil {
		if !strings.Contains(list, strings.ToUpper(vs.ValueStr)) {
			errors := errors.New(vs.Key + " must be in allowed list [ " + list + " ]")

			vs.OptionalData = errResponse
			vs.Error = &errors
		}

	}
	return vs
}

func (vs *StringValidatorContext) IsMustSameWith(allowedValue string, errResponse http.OptSetR) StringValidator {
	if vs.Error == nil && vs.ValueStr != allowedValue {
		errors := errors.New(vs.Key + " must be the same as the allowed value")

		vs.OptionalData = errResponse
		vs.Error = &errors

	}
	return vs
}

func (vs *StringValidatorContext) IsMustNotSameWith(disallowedValue string, errResponse http.OptSetR) StringValidator {
	if vs.Error == nil && vs.ValueStr == disallowedValue {
		errors := errors.New(vs.Key + " must not be the same as the disallowed value")

		vs.OptionalData = errResponse
		vs.Error = &errors

	}
	return vs
}

func (vs *StringValidatorContext) IsISO8601(errResponse http.OptSetR) StringValidator {
	if vs.Error == nil {
		_, err := time.Parse(time.RFC3339, vs.ValueStr)
		if err != nil {
			errors := errors.New(vs.Key + " must be a valid ISO 8601 date")

			vs.OptionalData = errResponse
			vs.Error = &errors
		}
	}
	return vs
}

func (vs *StringValidatorContext) IsAlphaNumeric(errResponse http.OptSetR) StringValidator {
	if vs.Error == nil {
		re := regexp.MustCompile("^[a-zA-Z0-9]+$")
		isAlphanumeric := re.MatchString(vs.ValueStr)
		if !isAlphanumeric {
			errors := errors.New(vs.Key + " must be in a alphanumeric format")

			vs.OptionalData = errResponse
			vs.Error = &errors
		}

	}
	return vs
}
