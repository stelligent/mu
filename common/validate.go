package common

import (
	"reflect"
	"regexp"

	"github.com/go-validator/validator"
)

// Validate validates the config struct
func (config *Config) Validate() error {
	validators()
	return validator.Validate(config)
}

// Validators registers the custom validators with the default validator
func validators() {
	validator.SetValidationFunc("emptyRegexp", emptyRegexp)
}

func emptyRegexp(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	value := st.String()
	if value == "" {
		return nil
	}
	return regex(value, param)
}

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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
