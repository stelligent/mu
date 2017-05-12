package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/pkg/errors"
	"strings"
)

// TaskContainerLister for listing tasks with containers
type TaskContainerLister interface {
	ListTasks(environment string, serviceName string) ([]Task, error)
}

// TaskCommandExecutor for executing commands against an environment
type TaskCommandExecutor interface {
	ExecuteCommand(task Task) (ECSRunTaskResult, error)
}

// TaskManager composite of all task capabilities
type TaskManager interface {
	TaskContainerLister
	TaskCommandExecutor
}

type ecsTaskManager struct {
	ec2API       ec2iface.EC2API
	ecsAPI       ecsiface.ECSAPI
	stackManager StackGetter
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

// NewTaskManager need for testing
func NewTaskManager(ec2API ec2iface.EC2API, ecsAPI ecsiface.ECSAPI, stackManager StackGetter) (TaskManager, error) {
	return &ecsTaskManager{
		ec2API:       ec2API,
		ecsAPI:       ecsAPI,
		stackManager: stackManager,
	}, nil
}

func newTaskManager(sess *session.Session, dryRun bool) (TaskManager, error) {
	log.Debug(EcsConnectionLog)

	ecsAPI := ecs.New(sess)
	ec2API := ec2.New(sess)
	stackManager, err := newStackManager(sess, dryRun)
	if err != nil {
		return nil, err
	}

	return &ecsTaskManager{
		ec2API:       ec2API,
		ecsAPI:       ecsAPI,
		stackManager: stackManager,
	}, nil
}

func (taskMgr *ecsTaskManager) getTaskRunInput(task Task) (*ecs.RunTaskInput, error) {
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

	ecsRunTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(ecsCluster),
		TaskDefinition: aws.String(ecsTaskDefinition),
		Count:          aws.Int64(ECSRunTaskDefaultCount),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				{
					Name: aws.String(ecsServiceName),
					Command: []*string{
						aws.String(strings.TrimSpace(task.Command)),
					},
				},
			},
		},
	}
	log.Debugf(ExecuteECSInputContentsLog, ecsRunTaskInput)
	return ecsRunTaskInput, nil
}

func (taskMgr *ecsTaskManager) runTask(runTaskInput *ecs.RunTaskInput) (ECSRunTaskResult, error) {
	resp, err := taskMgr.ecsAPI.RunTask(runTaskInput)
	log.Debugf(ExecuteECSResultContentsLog, resp, err)
	log.Info(ExecuteCommandFinishLog)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExecuteCommand runs a command for a specific environment
func (taskMgr *ecsTaskManager) ExecuteCommand(task Task) (ECSRunTaskResult, error) {
	log.Infof(ExecuteCommandStartLog, task.Command, task.Environment, task.Service)

	ecsRunTaskInput, err := taskMgr.getTaskRunInput(task)
	if err != nil {
		return nil, err
	}

	return taskMgr.runTask(ecsRunTaskInput)
}

// ExecuteCommand runs a command for a specific environment
func (taskMgr *ecsTaskManager) ListTasks(environment string, serviceName string) ([]Task, error) {
	cluster := CreateStackName(StackTypeCluster, environment)
	serviceInputParameters := &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	}
	tasks := []Task{}

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

func getTaskDetail(ecsTask *ecs.Task, taskMgr *ecsTaskManager, cluster string, environment string, serviceName string) (*Task, error) {
	log.Debugf(SvcGetTaskInfoLog, *ecsTask.TaskArn)
	containers := []Container{}
	if len(ecsTask.Containers) > Zero {
		for _, container := range ecsTask.Containers {
			if *container.Name != serviceName && len(serviceName) != Zero {
				return nil, errors.New(Empty)
			}
			containers = append(containers, getContainer(taskMgr, cluster, *ecsTask.ContainerInstanceArn, *container))
		}
	}
	task := Task{
		Name:        strings.Split(*ecsTask.TaskArn, TaskARNSeparator)[TaskGUIDIndex],
		Environment: environment,
		Service:     serviceName,
		Containers:  containers,
	}
	log.Debugf(SvcTaskDetailLog, task)
	return &task, nil
}

func getContainer(taskMgr *ecsTaskManager, cluster string, instanceARN string, container ecs.Container) Container {
	containerParams := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: []*string{
			aws.String(instanceARN),
		},
		Cluster: aws.String(cluster),
	}
	instanceOutput, err := taskMgr.ecsAPI.DescribeContainerInstances(containerParams)
	if err != nil {
		return Container{}
	}
	ec2InstanceID := *instanceOutput.ContainerInstances[FirstValueIndex].Ec2InstanceId
	ipAddress := getInstancePrivateIPAddress(taskMgr.ec2API, ec2InstanceID)
	return Container{Name: *container.Name, Instance: ec2InstanceID, PrivateIP: ipAddress}
}

func getInstancePrivateIPAddress(ec2API ec2iface.EC2API, instanceID string) string {
	ec2InputParameters := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}
	ec2Details, err := ec2API.DescribeInstances(ec2InputParameters)
	if err != nil {
		return Empty
	}
	ipAddress := *ec2Details.Reservations[FirstValueIndex].Instances[FirstValueIndex].PrivateIpAddress
	log.Debugf(SvcInstancePrivateIPLog, instanceID, ipAddress)

	return ipAddress
}

func (taskMgr *ecsTaskManager) getECSStack(serviceName string, environment string) (*Stack, error) {
	envStackName := CreateStackName(StackTypeService, serviceName, environment)
	log.Infof(SvcCmdStackLog, envStackName)

	return taskMgr.stackManager.GetStack(envStackName)
}
