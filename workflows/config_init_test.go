package workflows

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewConfigInitializer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	lister := NewConfigInitializer(ctx, false, 8080, false)
	assert.NotNil(lister)
}

func TestSubstituteEnvironmentVariables(t *testing.T) {
	assert := assert.New(t)
	input := `
---
environments:
  - name: prefix/${env:LOGNAME}/suffix
  - home: prefix/${env:HOME}/suffix
  - shell: prefix/${env:SHELL}/suffix
  - junk: prejunk/${env:junkymcjunkface}/postjunk
`
	output := common.SubstituteEnvironmentVariable(input)

	assert.NotContains(output, "LOGNAME")
	assert.Contains(output, "prefix/"+os.Getenv("LOGNAME")+"/suffix")

	assert.NotContains(output, "HOME")
	assert.Contains(output, "prefix/"+os.Getenv("HOME")+"/suffix")

	assert.NotContains(output, "SHELL")
	assert.Contains(output, "prefix/"+os.Getenv("SHELL")+"/suffix")

	// this variable should never exist, and should be replaced with nothing
	assert.NotContains(output, "junkymcjunkface")
	assert.Contains(output, "prejunk//postjunk")

}

func TestNewConfigInitializer_FileExists(t *testing.T) {
	assert := assert.New(t)

	var err error
	config := new(common.Config)
	config.Repo.Slug = "foo/bar"
	config.Basedir, err = ioutil.TempDir("", "mu-test")
	defer os.RemoveAll(config.Basedir)

	workflow := new(configWorkflow)
	err = workflow.configInitialize(config, false, 80, false)()
	assert.Nil(err)

	if newConfig, err := loadConfig(config.Basedir); err == nil {
		assert.Equal(80, newConfig.Service.Port)
	} else {
		assert.Fail(err.Error())
	}

	err = workflow.configInitialize(config, false, 80, false)()
	assert.NotNil(err)

	err = workflow.configInitialize(config, false, 3000, true)()
	assert.Nil(err)

	if newConfig, err := loadConfig(config.Basedir); err == nil {
		assert.Equal(3000, newConfig.Service.Port)
	} else {
		assert.Fail(err.Error())
	}
}

func loadConfig(basedir string) (*common.Config, error) {
	config := new(common.Config)
	yamlFile, err := os.Open(fmt.Sprintf("%s/mu.yml", basedir))
	if err != nil {
		return nil, err
	}
	yamlReader := bufio.NewReader(yamlFile)
	yamlBuffer := new(bytes.Buffer)
	yamlBuffer.ReadFrom(yamlReader)
	return config, yaml.Unmarshal(yamlBuffer.Bytes(), config)
}
