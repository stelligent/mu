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
	ctx.Config.Repo.Name = "my-repo"
	ctx.Config.Repo.Slug = "foo/my-repo"

	err := workflow.serviceFinder("", ctx)()
	assert.Nil(err)
	assert.NotNil(workflow.pipelineConfig)
	assert.Equal("my-repo", workflow.serviceName)
	assert.Equal("foo/my-repo", workflow.pipelineConfig.Source.Repo)
	assert.Equal("GitHub", workflow.pipelineConfig.Source.Provider)

	ctx.Config.Service.Name = "my-service"
	ctx.Config.Service.Pipeline.Source.Provider = "CodeCommit"
	ctx.Config.Service.Pipeline.Source.Repo = "bar/my-repo"
	err = workflow.serviceFinder("", ctx)()
	assert.Nil(err)
	assert.NotNil(workflow.pipelineConfig)
	assert.Equal("my-service", workflow.serviceName)
	assert.Equal("bar/my-repo", workflow.pipelineConfig.Source.Repo)
	assert.Equal("CodeCommit", workflow.pipelineConfig.Source.Provider)
}
