package common

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewContext(t *testing.T) {
	assert := assert.New(t)

	context := NewContext()

	assert.NotNil(context)
}

func TestLoadYamlConfig(t *testing.T) {
	assert := assert.New(t)

	yamlConfig :=
		`
---
environments:
  - name: dev
    loadbalancer:
      hostedzone: api-dev.example.com
    cluster:
      desiredCapacity: 1
      maxSize: 1
  - name: production
    loadbalancer:
      hostedzone: api.example.com
    cluster:
      desiredCapacity: 2
      maxSize: 5
service:
  desiredCount: 2
`

	context := NewContext()
	config := &context.Config
	err := loadYamlConfig(config, strings.NewReader(yamlConfig))

	assert.Nil(err)

	assert.NotNil(config)
	assert.Equal(2, len(config.Environments))
	assert.Equal("dev", config.Environments[0].Name)
	assert.Equal("api-dev.example.com", config.Environments[0].Loadbalancer.HostedZone)
	assert.Equal(1, config.Environments[0].Cluster.DesiredCapacity)
	assert.Equal(1, config.Environments[0].Cluster.MaxSize)
	assert.Equal("production", config.Environments[1].Name)
	assert.Equal("api.example.com", config.Environments[1].Loadbalancer.HostedZone)
	assert.Equal(2, config.Environments[1].Cluster.DesiredCapacity)
	assert.Equal(5, config.Environments[1].Cluster.MaxSize)

	assert.Equal(2, config.Service.DesiredCount)

}

func TestLoadBadYamlConfig(t *testing.T) {
	assert := assert.New(t)

	yamlConfig := `   blah blah blah   `

	context := NewContext()
	config := &context.Config
	err := loadYamlConfig(config, strings.NewReader(yamlConfig))
	assert.NotNil(err)
}
