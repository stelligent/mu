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
	ExecuteCommand(environmentName string, command string, service string) (ECSRunTaskResult, error)
}

// TaskManager composite of all task capabilities
type TaskManager interface {
	TaskCommandExecutor
}

type ecsTaskManager struct {
	ecsAPI       ecsiface.ECSAPI
	stackManager StackManager
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

// ExecuteCommand runs a command for a specific environment
func (taskMgr *ecsTaskManager) ExecuteCommand(environment string, service string, command string) (ECSRunTaskResult, error) {
	log.Infof(ExecuteCommandStartLog, command, environment, service)
	stackManager, err := taskMgr.stackManager.GetStack(service)
	if err != nil {
		return nil, err
	}
	ecsServiceName := stackManager.Parameters[ECSServiceNameParameterKey]
	ecsTaskDefinitionName := stackManager.Outputs[ECSTaskDefinitionOutputKey]

	ecsRunTaskInput := &ecs.RunTaskInput{
		TaskDefinition: aws.String(ecsTaskDefinitionName),
		Count:          aws.Int64(ECSRunTaskDefaultCount),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				{
					Name: aws.String(ecsServiceName),
					Command: []*string{
						aws.String(strings.TrimSpace(command)),
					},
				},
			},
		},
	}

	log.Debugf(ExecuteECSInputContentsLog, ecsRunTaskInput)

	resp, err := taskMgr.ecsAPI.RunTask(ecsRunTaskInput)
	log.Debugf(ExecuteECSResultContentsLog, resp, err)
	log.Info(ExecuteCommandFinishLog)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
