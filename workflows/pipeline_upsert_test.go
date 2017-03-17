package workflows

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewPipelineUpserter(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	upserter := NewPipelineUpserter(ctx, nil)
	assert.NotNil(upserter)
}

func TestPipelineBucket(t *testing.T) {
	assert := assert.New(t)

	workflow := new(pipelineWorkflow)
	workflow.serviceName = "my-service"

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-bucket-codepipeline").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-bucket-codepipeline", mock.AnythingOfType("map[string]string")).Return(nil)

	err := workflow.pipelineBucket(stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

	stackParams := stackManager.Calls[0].Arguments.Get(1).(map[string]string)
	assert.NotNil(stackParams)
	assert.Equal("codepipeline", stackParams["BucketPrefix"])
}

func TestPipelineUpserter(t *testing.T) {
	assert := assert.New(t)

	workflow := new(pipelineWorkflow)
	workflow.serviceName = "my-service"
	workflow.pipelineConfig = new(common.Pipeline)
	workflow.pipelineConfig.Source.Repo = "foo/bar"
	workflow.pipelineConfig.Source.Provider = "GitHub"

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-pipeline-my-service").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-pipeline-my-service", mock.AnythingOfType("map[string]string")).Return(nil)

	tokenProvider := func(required bool) string {
		return "my-token"
	}

	err := workflow.pipelineUpserter(tokenProvider, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

	stackParams := stackManager.Calls[1].Arguments.Get(1).(map[string]string)
	assert.NotNil(stackParams)
	assert.Equal("foo/bar", stackParams["SourceRepo"])
	assert.Equal("", stackParams["Branch"])
	assert.Equal("my-token", stackParams["GitHubToken"])
}
