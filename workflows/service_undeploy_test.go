package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServiceUndeployer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	terminator := NewServiceUndeployer(ctx, "foo", "dev")
	assert.NotNil(terminator)
}

func TestServiceUndeployer(t *testing.T) {
	assert := assert.New(t)

	workflow := new(serviceWorkflow)
	workflow.serviceName = "foo"

	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-service-foo-dev").Return(&common.Stack{Status: common.StackStatusDeleteComplete})
	stackManager.On("DeleteStack", "mu-service-foo-dev").Return(nil)

	err := workflow.serviceUndeployer("mu", "dev", stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	stackManager.AssertNumberOfCalls(t, "DeleteStack", 1)
}
