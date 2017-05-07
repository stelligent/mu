package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"strings"
)

// TaskCommandExecutor for executing commands against an environment
type TaskCommandExecutor interface {
	ExecuteCommand(task Task) (ECSRunTaskResult, error)
}

// TaskManager composite of all task capabilities
type TaskManager interface {
	TaskCommandExecutor
}

type ecsTaskManager struct {
	ecsAPI       ecsiface.ECSAPI
	stackManager StackManager
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

func newTaskManager(sess *session.Session, dryRun bool) (TaskManager, error) {
	log.Debug(EcsConnectionLog)
	ecsAPI := ecs.New(sess)
	stackManager, err := newStackManager(sess, dryRun)
	if err != nil {
		return nil, err
	}

	return &ecsTaskManager{
		ecsAPI:       ecsAPI,
		stackManager: stackManager,
	}, nil
}

func getTaskRunInput(stackManager StackManager, task Task) (*ecs.RunTaskInput, error) {
	envStackName := CreateStackName(StackTypeService, task.Service, task.Environment)
	log.Infof(SvcCmdStackLog, envStackName)

	ecsStack, err := stackManager.GetStack(envStackName)
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

	ecsRunTaskInput, err := getTaskRunInput(taskMgr.stackManager, task)
	if err != nil {
		return nil, err
	}

	return taskMgr.runTask(ecsRunTaskInput)
}
