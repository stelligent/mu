package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/pkg/errors"
	"github.com/stelligent/mu/common"
	"strings"
)

type ecsTaskManager struct {
	ecsAPI       ecsiface.ECSAPI
	stackManager common.StackGetter
}

func getFlagOrValue(flag string, value string) string {
	var actual string
	if len(flag) == Zero {
		actual = value
	} else {
		actual = flag
	}
	return actual
}

func newTaskManager(sess *session.Session, stackManager *common.StackManager) (common.TaskManager, error) {
	log.Debug(EcsConnectionLog)

	ecsAPI := ecs.New(sess)

	return &ecsTaskManager{
		ecsAPI:       ecsAPI,
		stackManager: *stackManager,
	}, nil
}

func (taskMgr *ecsTaskManager) getTaskRunInput(task common.Task) (*ecs.RunTaskInput, error) {
	ecsStack, err := taskMgr.getECSStack(task.Service, task.Environment)
	if err != nil {
		return nil, err
	}

	taskDefinitionOutput := ecsStack.Outputs[ECSTaskDefinitionOutputKey]
	ecsClusterOutput := ecsStack.Outputs[ECSClusterOutputKey]
	ecsTaskDefinition := getFlagOrValue(task.TaskDefinition, taskDefinitionOutput)
	ecsCluster := getFlagOrValue(task.Cluster, ecsClusterOutput)
	ecsServiceName := ecsStack.Parameters[ECSServiceNameParameterKey]
	log.Debugf(ExecuteECSInputParameterLog, task.Environment, ecsServiceName, ecsCluster, ecsTaskDefinition)

	command := make([]*string, len(task.Command))
	for i, commandPart := range task.Command {
		command[i] = aws.String(strings.TrimSpace(commandPart))
	}

	ecsRunTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(ecsCluster),
		TaskDefinition: aws.String(ecsTaskDefinition),
		Count:          aws.Int64(1),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				{
					Name:    aws.String(ecsServiceName),
					Command: command,
				},
			},
		},
	}
	log.Debugf(ExecuteECSInputContentsLog, ecsRunTaskInput)
	return ecsRunTaskInput, nil
}

func (taskMgr *ecsTaskManager) runTask(runTaskInput *ecs.RunTaskInput) (common.ECSRunTaskResult, error) {
	resp, err := taskMgr.ecsAPI.RunTask(runTaskInput)
	log.Debugf(ExecuteECSResultContentsLog, resp, err)
	log.Info(ExecuteCommandFinishLog)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExecuteCommand runs a command for a specific environment
func (taskMgr *ecsTaskManager) ExecuteCommand(task common.Task) (common.ECSRunTaskResult, error) {
	log.Infof(ExecuteCommandStartLog, task.Command, task.Environment, task.Service)

	ecsRunTaskInput, err := taskMgr.getTaskRunInput(task)
	if err != nil {
		return nil, err
	}

	return taskMgr.runTask(ecsRunTaskInput)
}

// ExecuteCommand runs a command for a specific environment
func (taskMgr *ecsTaskManager) ListTasks(environment string, serviceName string) ([]common.Task, error) {
	cluster := common.CreateStackName(common.StackTypeEnv, environment)
	serviceInputParameters := &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	}
	tasks := []common.Task{}

	serviceOutput, err := taskMgr.ecsAPI.ListServices(serviceInputParameters)
	if err != nil {
		return nil, err
	}

	for _, serviceARN := range serviceOutput.ServiceArns {
		log.Debugf(SvcListTasksLog, environment, cluster, serviceName)
		listTaskInput := &ecs.ListTasksInput{
			Cluster:     aws.String(cluster),
			ServiceName: aws.String(*serviceARN),
		}
		listTaskOutput, err := taskMgr.ecsAPI.ListTasks(listTaskInput)
		if err != nil {
			return nil, err
		}

		for _, ecsTask := range listTaskOutput.TaskArns {
			describeTaskParams := &ecs.DescribeTasksInput{
				Tasks: []*string{
					aws.String(*ecsTask),
				},
				Cluster: aws.String(cluster),
			}
			taskOutput, err := taskMgr.ecsAPI.DescribeTasks(describeTaskParams)
			if err != nil {
				continue
			}

			for _, ecsTask := range taskOutput.Tasks {
				task, err := getTaskDetail(ecsTask, taskMgr, cluster, environment, serviceName)
				if err != nil {
					continue
				}
				tasks = append(tasks, *task)
			}
		}
	}
	return tasks, nil
}

func getTaskDetail(ecsTask *ecs.Task, taskMgr *ecsTaskManager, cluster string, environment string, serviceName string) (*common.Task, error) {
	log.Debugf(SvcGetTaskInfoLog, *ecsTask.TaskArn)
	containers := []common.Container{}
	if len(ecsTask.Containers) > Zero {
		for _, container := range ecsTask.Containers {
			if *container.Name != serviceName && len(serviceName) != Zero {
				return nil, errors.New(common.Empty)
			}
			containers = append(containers, getContainer(taskMgr, cluster, *ecsTask.ContainerInstanceArn, *container))
		}
	}
	task := common.Task{
		Name:        strings.Split(*ecsTask.TaskArn, TaskARNSeparator)[1],
		Environment: environment,
		Service:     serviceName,
		Containers:  containers,
	}
	log.Debugf(SvcTaskDetailLog, task)
	return &task, nil
}

func getContainer(taskMgr *ecsTaskManager, cluster string, instanceARN string, container ecs.Container) common.Container {
	containerParams := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{
			aws.String(instanceARN),
		},
		Cluster: aws.String(cluster),
	}
	instanceOutput, err := taskMgr.ecsAPI.DescribeContainerInstances(containerParams)
	if err != nil {
		return common.Container{}
	}
	ec2InstanceID := *instanceOutput.ContainerInstances[0].Ec2InstanceId
	return common.Container{Name: *container.Name, Instance: ec2InstanceID}
}

func (taskMgr *ecsTaskManager) getECSStack(serviceName string, environment string) (*common.Stack, error) {
	envStackName := common.CreateStackName(common.StackTypeService, serviceName, environment)
	log.Infof(SvcCmdStackLog, envStackName)

	return taskMgr.stackManager.GetStack(envStackName)
}
