package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfigServicePort(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configMax := Config{
		Service: Service{
			Port: 65536,
		},
	}
	config := Config{
		Service: Service{
			Port: 2,
		},
	}

	assert.NotNil(configMax.Validate())
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}

func TestValidateConfigNamespace(t *testing.T) {
	assert := assert.New(t)

	configEmpty := Config{}
	configDash := Config{
		Namespace: "-invalid",
	}
	configNumeric := Config{
		Namespace: "0invalid",
	}
	config := Config{
		Namespace: "c00l-stack-name",
	}

	assert.NotNil(configDash.Validate())
	assert.NotNil(configNumeric.Validate())
	assert.Nil(configEmpty.Validate())
	assert.Nil(config.Validate())
}

func validateConfig(yamlConfig string) error {
	context := NewContext()
	config := &context.Config
	loadYamlConfig(config, strings.NewReader(yamlConfig))
	return config.Validate()
}
