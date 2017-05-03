package workflows

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewDatabaseUpserter(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	deploy := NewDatabaseUpserter(ctx, "dev")
	assert.NotNil(deploy)
}

type mockedRdsManager struct {
	mock.Mock
}

func (m *mockedRdsManager) SetIamAuthentication(dbInstanceName string, enabled bool, dbEngine string) error {
	args := m.Called(dbInstanceName)
	return args.Error(0)
}

func TestDatabaseUpserter_NoName(t *testing.T) {
	assert := assert.New(t)

	stackManager := new(mockedStackManagerForService)
	rdsManager := new(mockedRdsManager)

	config := new(common.Config)
	config.Service.Name = "foo"

	params := make(map[string]string)

	workflow := new(databaseWorkflow)
	workflow.serviceName = "foo"
	err := workflow.databaseDeployer(&config.Service, params, "dev", stackManager, stackManager, rdsManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 0)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 0)

	rdsManager.AssertExpectations(t)
	rdsManager.AssertNumberOfCalls(t, "SetIamAuthentication", 0)

}

func TestDatabaseUpserter(t *testing.T) {
	assert := assert.New(t)

	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-database-foo-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-database-foo-dev").Return(nil)

	rdsManager := new(mockedRdsManager)
	rdsManager.On("SetIamAuthentication", mock.Anything).Return(nil)

	config := new(common.Config)
	config.Service.Name = "foo"
	config.Service.Database.Name = "foo"

	params := make(map[string]string)

	workflow := new(databaseWorkflow)
	workflow.serviceName = "foo"
	err := workflow.databaseDeployer(&config.Service, params, "dev", stackManager, stackManager, rdsManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

	rdsManager.AssertExpectations(t)
	rdsManager.AssertNumberOfCalls(t, "SetIamAuthentication", 1)

}
