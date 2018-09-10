package common

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-validator/validator"
)

// Validate validates the config struct
func (config *Config) Validate() error {
	validators()
	return validator.Validate(config)
}

// Validators registers the custom validators with the default validator
func validators() {
	validator.SetValidationFunc("validateRoleARN", validateRoleARN)
	validator.SetValidationFunc("validateLeadingAlphaNumericDash", validateLeadingAlphaNumericDash)
	validator.SetValidationFunc("validateAlphaNumericDash", validateAlphaNumericDash)
	validator.SetValidationFunc("validateResourceID", validateResourceID)
	validator.SetValidationFunc("validateURL", validateURL)
	validator.SetValidationFunc("validateInstanceType", validateInstanceType)
	validator.SetValidationFunc("validateCIDR", validateCIDR)
	validator.SetValidationFunc("validateDockerImage", validateDockerImage)
}

func validateResourceID(v interface{}, param string) error {
	// TODO: Validate length of id - uses 8 or 17 character id
	st := reflect.ValueOf(v)
	kind := st.Kind().String()
	if kind == "string" {
		value := st.String()
		if value == "" {
			return nil
		}
		pattern := strings.Join([]string{"^", param, "-[a-zA-Z0-9]+$"}, "")
		return regex(value, pattern)
	}
	if kind == "slice" {
		// return some(st, pattern)
		return someString(st, param, validateResourceID)
	}
	return validator.ErrBadParameter
}

func isSlice(v interface{}) (reflect.Value, error) {
	st := reflect.ValueOf(v)
	kind := st.Kind().String()
	if kind == "slice" {
		return st, nil
	}
	return st, errors.New("not a slice")
}

func validateCIDR(v interface{}, param string) error {
	pattern := "^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}/\\d{1,2}$"
	return regexpLength(reflect.ValueOf(v).String(), pattern, 18)
}

func validateDockerImage(v interface{}, param string) error {
	// TODO: Validate Docker Image registry/image:tag
	return nil
}

// validateRoleARN validates that the value is an valid role ARN
func validateRoleARN(v interface{}, param string) error {
	value := reflect.ValueOf(v).String()
	pattern := "^arn:aws:iam::[0-9]{12}:role\\/[a-zA-Z0-9-+=\\/,.@_]+$"
	return regexpLength(value, pattern, 95)
}

// validateInstanceType validates the value is an instance type https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html
func validateInstanceType(v interface{}, param string) error {
	value := reflect.ValueOf(v).String()
	pattern := "^[a-zA-Z0-9]{2,3}\\.([a-zA-Z0-9]{2,3}\\.)?[a-zA-Z0-9]{4,10}$"
	return regexpLength(value, pattern, 95)
}

// validateURL validates that the string is a valid http resource
func validateURL(v interface{}, param string) error {
	value := reflect.ValueOf(v).String()
	pattern := "^[a-zA-Z0-9/][a-zA-Z0-9-\\./_]*?$"
	return regexpLength(value, pattern, 255)
}

// validateLeadingAlphaNumericDash checks for alphanumric strings with a dash that starts with an alphanumeric character
func validateLeadingAlphaNumericDash(v interface{}, param string) error {
	value := reflect.ValueOf(v).String()
	pattern := "^[a-zA-Z0-9][a-zA-Z0-9-]+$"
	// default length for stackName
	length := 63
	if p, _ := strconv.Atoi(param); p != 0 {
		length = p
	}
	return regexpLength(value, pattern, length)
}

// validateAlphaNumericDash is similar to validateLeadingAlphaNumericDash but requires starting alpha character
func validateAlphaNumericDash(v interface{}, param string) error {
	value := reflect.ValueOf(v).String()
	pattern := "^[a-zA-Z][a-zA-Z0-9-]+$"
	// default length for stackName
	length := 63
	if p, _ := strconv.Atoi(param); p != 0 {
		length = p
	}
	return regexpLength(value, pattern, length)
}

func regexpLength(value string, pattern string, max int) error {
	if value == "" {
		return nil
	}
	if len(value) > max {
		return validator.ErrMax
	}
	return regex(value, pattern)
}

// someString checks a string array until one returns an error
func someString(s reflect.Value, param string, fun func(interface{}, string) error) error {
	for i := 0; i < s.Len(); i++ {
		if err := fun(s.Index(i).String(), param); err != nil {
			return err
		}
	}
	return nil
}

// regex checks if a string matches a regular express "string"
func regex(v string, param string) error {
	re, err := regexp.Compile(param)
	if err != nil {
		return validator.ErrBadParameter
	}

	if !re.MatchString(v) {
		return validator.ErrRegexp
	}
	return nil
}
