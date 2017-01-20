package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceLoader(t *testing.T) {
	assert := assert.New(t)

	config := new(common.Config)
	config.Service.Name = "myservice"

	workflow := new(serviceWorkflow)
	err := workflow.serviceLoader(config)()
	assert.Nil(err)
	assert.Equal("myservice", workflow.service.Name)
}
