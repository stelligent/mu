package common

import (
	"testing"
	"github.com/stretchr/testify/assert"

)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	config := NewConfig()

	assert.NotNil(config)
}

func TestLoadYamlConfig(t *testing.T) {
	assert := assert.New(t)

	yamlConfig :=
`
---
environments:
  - name: dev
    loadbalancer:
      hostname: api-dev.example.com
    cluster:
      desiredCapacity: 1
      maxSize: 1
  - name: production
    loadbalancer:
      hostname: api.example.com
    cluster:
      desiredCapacity: 2
      maxSize: 5
service:
  desiredCount: 2
`


	config := NewConfig()
	config.loadFromYaml([]byte(yamlConfig))

	assert.NotNil(config)
	assert.Equal(2,len(config.Environments))
	assert.Equal("dev",config.Environments[0].Name)
	assert.Equal("api-dev.example.com",config.Environments[0].Loadbalancer.Hostname)
	assert.Equal(1,config.Environments[0].Cluster.DesiredCapacity)
	assert.Equal(1,config.Environments[0].Cluster.MaxSize)
	assert.Equal("production",config.Environments[1].Name)
	assert.Equal("api.example.com",config.Environments[1].Loadbalancer.Hostname)
	assert.Equal(2,config.Environments[1].Cluster.DesiredCapacity)
	assert.Equal(5,config.Environments[1].Cluster.MaxSize)

	assert.Equal(2,config.Service.DesiredCount)

}

func TestLoadBadYamlConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	yamlConfig := `   blah blah blah   `

	config := NewConfig()
	config.loadFromYaml([]byte(yamlConfig))
}

func TestLoadConfig(t *testing.T) {

	config := NewConfig()
	config.LoadFromFile("foobarbaz.yml")
}

func TestGetEnvironment(t *testing.T) {
	assert := assert.New(t)
	config := NewConfig()

	env1 := Environment{
		Name: "foo",
	}
	env2 := Environment{
		Name: "bar",
	}
	config.Environments = []Environment{env1, env2}

	fooEnv, fooErr := config.GetEnvironment("foo")
	assert.Equal("foo", fooEnv.Name)
	assert.Nil(fooErr)

	barEnv, barErr := config.GetEnvironment("BAR")
	assert.Equal("bar", barEnv.Name)
	assert.Nil(barErr)

	bazEnv, bazErr := config.GetEnvironment("baz")
	assert.Nil(bazEnv)
	assert.NotNil(bazErr)
}
