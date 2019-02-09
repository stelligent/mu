package workflows

import (
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
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

	paramManager := new(mockedParamManager)
	paramManager.On("DeleteParam", "mu-database-foo-dev-DatabaseMasterPassword").Return(nil)

	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-database-foo-dev").Return(&common.Stack{Status: common.StackStatusDeleteComplete})
	stackManager.On("DeleteStack", "mu-database-foo-dev").Return(nil)

	err := workflow.databaseTerminator("mu", "dev", stackManager, stackManager, paramManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	stackManager.AssertNumberOfCalls(t, "DeleteStack", 1)
}
