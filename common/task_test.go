package common

import (
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockedStackManager struct {
	mock.Mock
	StackManager
}

func (m *mockedStackManager) GetStack(stackName string) (*Stack, error) {
	args := m.Called()
	return args.Get(0).(*Stack), args.Error(1)
}

func (m *mockedECS) RunTask(input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error) {
	args := m.Called()
	return args.Get(0).(*ecs.RunTaskOutput), args.Error(1)
}

func TestTaskCommandExecutor_succeed(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	ecsMock := new(mockedECS)

	stackManagerMock.On(GetStackName).Return(&Stack{}, nil)
	ecsMock.On(RunTaskName).Return(&ecs.RunTaskOutput{}, nil)

	executeManager := ecsTaskManager{
		ecsAPI:       ecsMock,
		stackManager: stackManagerMock,
	}
	task := Task{
		Environment: TestEnv,
		Service:     TestSvc,
		Command:     TestCmd,
	}
	result, err := executeManager.ExecuteCommand(task)
	assertion.NotNil(result)
	assertion.Nil(err)

	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
	ecsMock.AssertExpectations(t)
	ecsMock.AssertNumberOfCalls(t, RunTaskName, 1)
}
