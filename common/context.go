package common

import (
	"fmt"
	"io/ioutil"
)

// NewContext create a new context object
func NewContext() *Context {
	ctx := new(Context)
	return ctx
}

// InitializeFromFile loads config object from local file
func (ctx *Context) InitializeFromFile(configFile string) {
	yamlConfig, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("WARN: Unable to find config file: %v\n", err)
	} else {
		ctx.Config.loadFromYaml(yamlConfig)
	}

	ctx.Initialize()
}

// Initialize will create AWS services
func (ctx *Context) Initialize() error {
	cfn, err := newCloudFormation(ctx.Config.Region)
	if err != nil {
		return err
	}

	ctx.CloudFormation = cfn
	return nil
}
