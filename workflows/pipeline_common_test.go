package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceFinder(t *testing.T) {
	assert := assert.New(t)

	workflow := new(pipelineWorkflow)

	ctx := new(common.Context)
	ctx.Repo.Name = "my-repo"

	err := workflow.serviceFinder("", ctx)()
	assert.Nil(err)
	assert.NotNil(workflow.pipelineConfig)
	assert.Equal("my-repo", workflow.serviceName)

	ctx.Config.Service.Name = "my-service"
	err = workflow.serviceFinder("", ctx)()
	assert.Nil(err)
	assert.NotNil(workflow.pipelineConfig)
	assert.Equal("my-service", workflow.serviceName)
}
