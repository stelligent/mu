package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDatabaseTerminator(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	terminator := NewDatabaseTerminator(ctx, "foo", "dev")
	assert.NotNil(terminator)
}

func TestDatabaseTerminate(t *testing.T) {
	assert := assert.New(t)

	workflow := new(databaseWorkflow)
	workflow.serviceName = "foo"

	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-database-foo-dev").Return(&common.Stack{Status: common.StackStatusDeleteComplete})
	stackManager.On("DeleteStack", "mu-database-foo-dev").Return(nil)

	err := workflow.databaseTerminator("mu", "dev", stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	stackManager.AssertNumberOfCalls(t, "DeleteStack", 1)
}
