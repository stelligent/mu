package common

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"io"
)

// NewContext create a new context object
func NewContext() *Context {
	ctx := new(Context)
	return ctx
}

// Initialize loads config object
func (ctx *Context) Initialize(configReader io.Reader) error {
	// load the configuration
	err := loadYamlConfig(&ctx.Config, configReader)
	if err != nil {
		return err
	}

	// initialize StackManager
	ctx.StackManager, err = newStackManager(ctx.Config.Region)
	if err != nil {
		return err
	}

	return nil
}

func loadYamlConfig(config *Config, yamlReader io.Reader) error {
	yamlBuffer := new(bytes.Buffer)
	yamlBuffer.ReadFrom(yamlReader)
	return yaml.Unmarshal(yamlBuffer.Bytes(), config)
}
