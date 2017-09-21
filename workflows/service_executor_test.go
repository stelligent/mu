package workflows

import (
	"errors"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockedStackManager struct {
	mock.Mock
	common.StackManager
}

type mockedTaskManager struct {
	mock.Mock
	common.TaskManager
}

func (m *mockedStackManager) GetStack(stackName string) (*common.Stack, error) {
	args := m.Called()
	return args.Get(0).(*common.Stack), args.Error(1)
}

func (m *mockedTaskManager) ExecuteCommand(namespace string, task common.Task) (common.ECSRunTaskResult, error) {
	args := m.Called()
	return nil, args.Error(1)
}

func TestNewServiceExecutorCreate(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	executor := NewServiceExecutor(ctx, common.Task{})
	assertion.NotNil(executor)
}

func TestNewServiceExecutorFail(t *testing.T) {
	assertion := assert.New(t)
	taskManagerMock := new(mockedTaskManager)

	taskManagerMock.On("ExecuteCommand").Return(nil, errors.New(common.Empty))

	task := common.Task{
		Environment: TestEnv,
		Service:     TestSvc,
		Command:     []string{TestCmd},
	}
	executor := newServiceExecutor("mu", taskManagerMock, task)
	assertion.NotNil(executor)
	assertion.NotNil(executor())
}

func TestNewServiceExecutor(t *testing.T) {
	assertion := assert.New(t)
	taskManagerMock := new(mockedTaskManager)

	taskManagerMock.On("ExecuteCommand").Return(nil, nil)

	task := common.Task{
		Environment: TestEnv,
		Service:     TestSvc,
		Command:     []string{TestCmd},
	}
	executor := newServiceExecutor("mu", taskManagerMock, task)
	assertion.NotNil(executor)
	assertion.Nil(executor())

	taskManagerMock.AssertExpectations(t)
	taskManagerMock.AssertNumberOfCalls(t, "ExecuteCommand", 1)
}
