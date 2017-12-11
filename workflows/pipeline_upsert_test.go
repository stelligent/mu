package workflows

import (
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewPipelineUpserter(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	upserter := NewPipelineUpserter(ctx, nil)
	assert.NotNil(upserter)
}

func TestServiceBucketUpsert_CreateStack(t *testing.T) {
	assert := assert.New(t)

	workflow := new(pipelineWorkflow)
	workflow.serviceName = "my-service"
	workflow.pipelineConfig = new(common.Pipeline)

	bucketStack := &common.Stack{
		Status: common.StackStatusCreateComplete,
		Outputs: map[string]string{
			"Bucket": "foo-bucket",
		},
	}

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-bucket-codepipeline").Return(bucketStack)
	stackManager.On("UpsertStack", "mu-bucket-codepipeline", mock.AnythingOfType("map[string]string")).Return(nil)

	stackParams := make(map[string]string)

	err := workflow.pipelineBucket("mu", stackParams, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

	assert.NotNil(stackParams)
	assert.Equal("foo-bucket", stackParams["PipelineBucket"])
}

func TestPipelineBucket_Configured(t *testing.T) {
	assert := assert.New(t)

	workflow := new(pipelineWorkflow)
	workflow.serviceName = "my-service"
	workflow.pipelineConfig = new(common.Pipeline)
	workflow.pipelineConfig.Bucket = "mu-test-bucket"

	stackManager := new(mockedStackManagerForUpsert)

	stackParams := make(map[string]string)
	err := workflow.pipelineBucket("mu", stackParams, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 0)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 0)

	assert.NotNil(stackParams)
	assert.Equal("mu-test-bucket", stackParams["PipelineBucket"])
}

func TestPipelineUpserter(t *testing.T) {
	assert := assert.New(t)

	workflow := new(pipelineWorkflow)
	workflow.serviceName = "my-service"
	workflow.pipelineConfig = new(common.Pipeline)
	workflow.pipelineConfig.Source.Repo = "foo/bar"
	workflow.pipelineConfig.Source.Provider = "GitHub"

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-pipeline-my-service").Return(&common.Stack{Status: common.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-pipeline-my-service", mock.AnythingOfType("map[string]string")).Return(nil)

	tokenProvider := func(required bool) string {
		return "my-token"
	}

	params := make(map[string]string)
	err := workflow.pipelineUpserter("mu", tokenProvider, stackManager, stackManager, params)()
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
