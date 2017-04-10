package workflows

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewEnvironmentTerminator(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	terminator := NewEnvironmentTerminator(ctx, "foo")
	assert.NotNil(terminator)
}

type mockedStackManagerForTerminate struct {
	mock.Mock
}

func (m *mockedStackManagerForTerminate) AwaitFinalStatus(stackName string) *common.Stack {
	args := m.Called(stackName)
	return args.Get(0).(*common.Stack)
}
func (m *mockedStackManagerForTerminate) DeleteStack(stackName string) error {
	args := m.Called(stackName)
	return args.Error(0)
}
func (m *mockedStackManagerForTerminate) FindLatestImageID(pattern string) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestNewEnvironmentEcsTerminator(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}

	stackManager := new(mockedStackManagerForTerminate)
	stackManager.On("AwaitFinalStatus", "mu-cluster-foo").Return(&common.Stack{Status: cloudformation.StackStatusDeleteComplete})
	stackManager.On("DeleteStack", "mu-cluster-foo").Return(nil)

	err := workflow.environmentEcsTerminator("foo", stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "DeleteStack", 1)
}

func TestNewEnvironmentConsulTerminator(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}

	stackManager := new(mockedStackManagerForTerminate)
	stackManager.On("AwaitFinalStatus", "mu-consul-foo").Return(&common.Stack{Status: cloudformation.StackStatusDeleteComplete})
	stackManager.On("DeleteStack", "mu-consul-foo").Return(nil)

	err := workflow.environmentConsulTerminator("foo", stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "DeleteStack", 1)
}

func TestNewEnvironmentVpcTerminator(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}

	stackManager := new(mockedStackManagerForTerminate)
	stackManager.On("AwaitFinalStatus", "mu-target-foo").Return(&common.Stack{Status: cloudformation.StackStatusDeleteComplete})
	stackManager.On("AwaitFinalStatus", "mu-vpc-foo").Return(&common.Stack{Status: cloudformation.StackStatusDeleteComplete})
	stackManager.On("DeleteStack", "mu-target-foo").Return(nil)
	stackManager.On("DeleteStack", "mu-vpc-foo").Return(nil)

	err := workflow.environmentVpcTerminator("foo", stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	stackManager.AssertNumberOfCalls(t, "DeleteStack", 2)
}
