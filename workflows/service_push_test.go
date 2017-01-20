package workflows

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServicePusher(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	upserter := NewServicePusher(ctx, "foo")
	assert.NotNil(upserter)
}

func TestServiceRepoUpserter(t *testing.T) {
	assert := assert.New(t)

	workflow := new(serviceWorkflow)
	workflow.service = &common.Service{
		Name: "foo",
	}

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-repo-foo").Return(cloudformation.StackStatusCreateComplete)
	stackManager.On("UpsertStack", "mu-repo-foo").Return(nil)

	err := workflow.serviceRepoUpserter(stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestServiceBuild(t *testing.T) {
	assert := assert.New(t)

	workflow := new(serviceWorkflow)
	workflow.service = &common.Service{
		Name: "foo",
	}

	err := workflow.serviceBuilder("foo")()
	assert.Nil(err)
}

func TestServicePush(t *testing.T) {
	assert := assert.New(t)

	workflow := new(serviceWorkflow)
	workflow.service = &common.Service{
		Name: "foo",
	}

	err := workflow.servicePusher("foo")()
	assert.Nil(err)
}
