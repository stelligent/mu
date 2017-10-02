package aws

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockedStackManager struct {
	mock.Mock
	common.StackManager
}

func (m *mockedStackManager) GetStack(stackName string) (*common.Stack, error) {
	args := m.Called()
	return args.Get(0).(*common.Stack), args.Error(1)
}

func (m *mockedECS) RunTask(input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error) {
	args := m.Called()
	return args.Get(0).(*ecs.RunTaskOutput), args.Error(1)
}

func (m *mockedECS) ListServices(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
	args := m.Called()
	return args.Get(0).(*ecs.ListServicesOutput), args.Error(1)
}

func (m *mockedECS) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	args := m.Called()
	return args.Get(0).(*ecs.DescribeTasksOutput), args.Error(1)
}

func (m *mockedECS) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	args := m.Called()
	return args.Get(0).(*ecs.ListTasksOutput), args.Error(1)
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

	stackManagerMock.On(GetStackName).Return(&common.Stack{}, nil)
	ecsMock.On(RunTaskName).Return(&ecs.RunTaskOutput{}, nil)

	executeManager := ecsTaskManager{
		ecsAPI:       ecsMock,
		stackManager: stackManagerMock,
	}
	task := getTestTask()
	result, err := executeManager.ExecuteCommand("mu", task)
	assertion.NotNil(result)
	assertion.Nil(err)

	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
	ecsMock.AssertExpectations(t)
	ecsMock.AssertNumberOfCalls(t, RunTaskName, 1)
}

func TestTaskListViewSucceed(t *testing.T) {
	stackManagerMock := new(mockedStackManager)
	ecsMock := new(mockedECS)
	serviceList := []*string{aws.String(TestSvc)}
	stackManagerMock.On(GetStackName).Return(&common.Stack{}, nil)
	ecsMock.On(ListTasks).Return(&ecs.ListTasksOutput{TaskArns: serviceList}, nil)
	ecsMock.On(ListServices).Return(&ecs.ListServicesOutput{ServiceArns: serviceList}, nil)
	ecsMock.On(DescribeTasks).Return(&ecs.DescribeTasksOutput{Tasks: []*ecs.Task{{ContainerInstanceArn: aws.String(TestEnv), TaskArn: aws.String(TestTaskARN), Containers: []*ecs.Container{{ContainerArn: aws.String(TestCmd), Name: aws.String(TestEnv)}}}}}, nil)
	ecsMock.On(DescribeContainerInstances).Return(&ecs.DescribeContainerInstancesOutput{ContainerInstances: []*ecs.ContainerInstance{{Ec2InstanceId: aws.String(TestSvc), ContainerInstanceArn: aws.String(TestCmd)}}}, nil)

	executeManager := ecsTaskManager{
		ecsAPI:       ecsMock,
		stackManager: stackManagerMock,
	}

	tasks, err := executeManager.ListTasks("mu", TestEnv, Empty)
	assert.Nil(t, err)
	assert.NotNil(t, tasks)

	ecsMock.AssertExpectations(t)
	ecsMock.AssertNumberOfCalls(t, ListServices, 1)
	ecsMock.AssertNumberOfCalls(t, ListTasks, 1)
	ecsMock.AssertNumberOfCalls(t, DescribeTasks, 1)
	ecsMock.AssertNumberOfCalls(t, DescribeContainerInstances, 1)
}

func TestTaskCommandExecutorFailRun(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	ecsMock := new(mockedECS)

	stackManagerMock.On(GetStackName).Return(&common.Stack{}, nil)
	ecsMock.On(RunTaskName).Return(&ecs.RunTaskOutput{}, errors.New(Empty))

	executeManager := ecsTaskManager{
		ecsAPI:       ecsMock,
		stackManager: stackManagerMock,
	}
	task := getTestTask()
	result, err := executeManager.ExecuteCommand("mu", task)
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

	stackManagerMock.On(GetStackName).Return(&common.Stack{}, errors.New(Empty))
	taskManager := ecsTaskManager{
		stackManager: stackManagerMock,
	}

	task := getTestTask()

	result, err := taskManager.ExecuteCommand("mu", task)
	assertion.NotNil(err)
	assertion.Nil(result)
}

func TestRunInputSucceed(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	outputs := make(map[string]string)
	parameters := make(map[string]string)
	outputs[ECSClusterOutputKey] = "clusterKey"
	outputs[ECSTaskDefinitionOutputKey] = "taskKey"
	ecsStackName := common.CreateStackName("mu", common.StackTypeService, TestEnv, TestSvc)
	parameters[ECSServiceNameParameterKey] = ecsStackName
	stackManagerMock.On(GetStackName).Return(&common.Stack{Outputs: outputs, Parameters: parameters}, nil)
	executeManager := ecsTaskManager{
		ecsAPI:       nil,
		stackManager: stackManagerMock,
	}
	task := getTestTask()
	runInput, err := executeManager.getTaskRunInput("mu", task)
	assertion.NotNil(runInput)
	assertion.Nil(err)
	assertion.Equal(*runInput.Cluster, "clusterKey")
	assertion.Equal(*runInput.TaskDefinition, "taskKey")
	assertion.Equal(*runInput.Overrides.ContainerOverrides[Zero].Name, ecsStackName)
	assertion.Equal(*runInput.Overrides.ContainerOverrides[Zero].Command[Zero], TestCmd)

	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
}

func TestRunInputFail(t *testing.T) {
	assertion := assert.New(t)
	stackManagerMock := new(mockedStackManager)
	stackManagerMock.On(GetStackName).Return(&common.Stack{}, errors.New(Empty))
	task := getTestTask()
	executeManager := ecsTaskManager{
		ecsAPI:       nil,
		stackManager: stackManagerMock,
	}
	badInput, inputErr := executeManager.getTaskRunInput("mu", task)
	assertion.NotNil(inputErr)
	assertion.Nil(badInput)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, GetStackName, 1)
}

func getTestTask() common.Task {
	return common.Task{
		Environment: TestEnv,
		Service:     TestSvc,
		Command:     []string{TestCmd},
	}
}
