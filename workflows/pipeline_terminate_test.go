package workflows

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPipelineTerminator(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	terminator := NewPipelineTerminator(ctx, "foo")
	assert.NotNil(terminator)
}

func TestPipelineTerminator(t *testing.T) {
	assert := assert.New(t)

	workflow := new(pipelineWorkflow)
	workflow.serviceName = "foo"

	stackManager := new(mockedStackManagerForTerminate)
	stackManager.On("AwaitFinalStatus", "mu-pipeline-foo").Return(&common.Stack{Status: cloudformation.StackStatusDeleteComplete})
	stackManager.On("DeleteStack", "mu-pipeline-foo").Return(nil)

	err := workflow.pipelineTerminator(stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "DeleteStack", 1)
}
