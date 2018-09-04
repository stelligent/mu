package common

import (
	"github.com/go-validator/validator"
)

// Validate validates the config struct
func (config *Config) Validate() error {
	return validator.Validate(config)
}

// Validators registers the custom validators with the default validator
func Validators() {
}
