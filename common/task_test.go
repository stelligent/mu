package common

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
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

func TestOptionalFlags(t *testing.T) {
	assertion := assert.New(t)
	assertion.Equal(TestEnv, getFlagOrValue(Empty, TestEnv))
	assertion.Equal(TestEnv, getFlagOrValue(TestEnv, Empty))
	assertion.Equal(TestEnv, getFlagOrValue(TestEnv, TestSvc))
	assertion.Equal(TestSvc, getFlagOrValue(TestSvc, TestEnv))
	assertion.Empty(getFlagOrValue(Empty, Empty))
}

func TestTaskCommandExecutorSucceed(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	ecsMock := new(mockedECS)

	stackManagerMock.On(GetStackName).Return(&Stack{}, nil)
	ecsMock.On(RunTaskName).Return(&ecs.RunTaskOutput{}, nil)

	executeManager := ecsTaskManager{
		ecsAPI:       ecsMock,
		stackManager: stackManagerMock,
	}
	task := getTestTask()
	result, err := executeManager.ExecuteCommand(task)
	assertion.NotNil(result)
	assertion.Nil(err)

	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
	ecsMock.AssertExpectations(t)
	ecsMock.AssertNumberOfCalls(t, RunTaskName, 1)
}

func TestTaskCommandExecutorFailRun(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	ecsMock := new(mockedECS)

	stackManagerMock.On(GetStackName).Return(&Stack{}, nil)
	ecsMock.On(RunTaskName).Return(&ecs.RunTaskOutput{}, errors.New(Empty))

	executeManager := ecsTaskManager{
		ecsAPI:       ecsMock,
		stackManager: stackManagerMock,
	}
	task := getTestTask()
	result, err := executeManager.ExecuteCommand(task)
	assertion.NotNil(err)
	assertion.Nil(result)

	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
	ecsMock.AssertExpectations(t)
	ecsMock.AssertNumberOfCalls(t, RunTaskName, 1)
}

func TestTaskCommandExecutorFailStack(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)

	stackManagerMock.On(GetStackName).Return(&Stack{}, errors.New(Empty))
	sess, err := session.NewSession()
	taskManager, err := newTaskManager(sess, false)

	task := getTestTask()
	result, err := taskManager.ExecuteCommand(task)
	assertion.NotNil(err)
	assertion.Nil(result)
}

func TestRunInputSucceed(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	outputs := make(map[string]string)
	parameters := make(map[string]string)
	outputs[ECSClusterOutputKey] = ClusterFlag
	outputs[ECSTaskDefinitionOutputKey] = TaskFlag
	ecsStackName := CreateStackName(StackTypeService, TestEnv, TestSvc)
	parameters[ECSServiceNameParameterKey] = ecsStackName
	stackManagerMock.On(GetStackName).Return(&Stack{Outputs: outputs, Parameters: parameters}, nil)
	task := getTestTask()
	runInput, err := getTaskRunInput(stackManagerMock, task)
	assertion.NotNil(runInput)
	assertion.Nil(err)
	assertion.Equal(*runInput.Cluster, ClusterFlag)
	assertion.Equal(*runInput.TaskDefinition, TaskFlag)
	assertion.Equal(*runInput.Overrides.ContainerOverrides[Zero].Name, ecsStackName)
	assertion.Equal(*runInput.Overrides.ContainerOverrides[Zero].Command[Zero], TestCmd)

	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
}

func TestRunInputFail(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	stackManagerMock.On(GetStackName).Return(&Stack{}, errors.New(Empty))
	task := getTestTask()

	badInput, inputErr := getTaskRunInput(stackManagerMock, task)
	assertion.NotNil(inputErr)
	assertion.Nil(badInput)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
}

func getTestTask() Task {
	return Task{
		Environment: TestEnv,
		Service:     TestSvc,
		Command:     TestCmd,
	}
}
