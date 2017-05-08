package workflows

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockedStackManager struct {
	mock.Mock
	common.StackManager
}

type mockedECS struct {
	mock.Mock
	ecsiface.ECSAPI
}

func (m *mockedStackManager) GetStack(stackName string) (*common.Stack, error) {
	args := m.Called()
	return args.Get(0).(*common.Stack), args.Error(1)
}

func (m *mockedECS) RunTask(input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error) {
	args := m.Called()
	return args.Get(0).(*ecs.RunTaskOutput), args.Error(1)
}

func TestNewServiceExecutorCreate(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	executor := NewServiceExecutor(ctx, common.Task{})
	assertion.NotNil(executor)
}

func TestNewServiceExecutorFail(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	ecsMock := new(mockedECS)

	stackManagerMock.On(common.GetStackName).Return(&common.Stack{}, nil)
	ecsMock.On(common.RunTaskName).Return(&ecs.RunTaskOutput{}, errors.New(common.Empty))

	taskManager, err := common.NewTaskManager(ecsMock, stackManagerMock)
	assertion.Nil(err)
	assertion.NotNil(taskManager)
	task := common.Task{
		Environment: common.TestEnv,
		Service:     common.TestSvc,
		Command:     common.TestCmd,
	}
	executor := newServiceExecutor(taskManager, task)
	assertion.NotNil(executor)
	assertion.NotNil(executor())
}

func TestNewServiceExecutor(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	ecsMock := new(mockedECS)

	stackManagerMock.On(common.GetStackName).Return(&common.Stack{}, nil)
	ecsMock.On(common.RunTaskName).Return(&ecs.RunTaskOutput{}, nil)

	taskManager, err := common.NewTaskManager(ecsMock, stackManagerMock)
	assertion.Nil(err)
	assertion.NotNil(taskManager)
	task := common.Task{
		Environment: common.TestEnv,
		Service:     common.TestSvc,
		Command:     common.TestCmd,
	}
	executor := newServiceExecutor(taskManager, task)
	assertion.NotNil(executor)
	assertion.Nil(executor())

	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, common.GetStackName, 1)
	ecsMock.AssertExpectations(t)
	ecsMock.AssertNumberOfCalls(t, common.RunTaskName, 1)
}
