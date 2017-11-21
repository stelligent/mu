package common

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"bufio"
	"io"
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

func TestSubstituteEnvironmentVariablessAsStream(t *testing.T) {
	assert := assert.New(t)
	inputStream := strings.NewReader(`
---
environments:
  - name: prefix/${env:LOGNAME}/suffix
  - home: prefix/${env:HOME}/suffix
  - shell: prefix/${env:SHELL}/suffix
  - junk: prejunk/${env:junkymcjunkface}/postjunk
`)
	substStream := SubstituteEnvironmentVariablesAsStream(inputStream)

	outputBuffer, err := ioutil.ReadAll(substStream)
	if err != nil {
		assert.Fail("couldn't ReadAll on substStream: %v", err)
	}
	output := string(outputBuffer)

	log.Infof("output1: %v", output)

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

func TestLoadBadYamlConfig(t *testing.T) {
	assert := assert.New(t)

	yamlConfig := `   blah blah blah   `

	context := NewContext()
	config := &context.Config
	err := loadYamlConfig(config, strings.NewReader(yamlConfig))
	assert.NotNil(err)
}

func TestMainForTestingStreamProcessor(t *testing.T) {
	assert := assert.New(t)

	input := `
  - name: prefix/${env:LOGNAME}/suffix
  - home: prefix/${env:HOME}/suffix
  - shell: prefix/${env:SHELL}/suffix
  - junk: prejunk/${env:junkymcjunkface}/postjunk `

	var reader1 io.Reader = strings.NewReader(input)
	var reader2 *bufio.Reader = bufio.NewReader(reader1)
	p := &StreamProcessor{ LineReader: *reader2 }

	 outputBytes, err := ioutil.ReadAll(p)
	 if err != nil {
	 	log.Infof("error processing")
	 	os.Exit(1)
	 }
	 outputString := string(outputBytes)

	log.Infof("outputString '%v'", outputString)

	assert.NotContains(outputString, "LOGNAME")
	assert.Contains(outputString, "prefix/"+os.Getenv("LOGNAME")+"/suffix")

	assert.NotContains(outputString, "HOME")
	assert.Contains(outputString, "prefix/"+os.Getenv("HOME")+"/suffix")

	assert.NotContains(outputString, "SHELL")
	assert.Contains(outputString, "prefix/"+os.Getenv("SHELL")+"/suffix")

	// this variable should never exist, and should be replaced with nothing
	assert.NotContains(outputString, "junkymcjunkface")
	assert.Contains(outputString, "prejunk//postjunk")
}
