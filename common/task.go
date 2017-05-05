package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"os"
	"strings"
)

// TaskCommandExecutor for executing commands against an environment
type TaskCommandExecutor interface {
	ExecuteCommand(environmentName string, command string, service string) (string, error)
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
func (taskMgr *ecsTaskManager) ExecuteCommand(environment string, service string, command string) (string, error) {
	stack, err := taskMgr.stackManager.GetStack(service)
	if err != nil {
		return Empty, err
	}
	ecsServiceName := stack.Parameters[ECSServiceNameParameterKey]
	ecsTaskDefinitionName := stack.Outputs[ECSTaskDefinitionOutputKey]

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
	fmt.Fprintf(os.Stdout, "Executing command '[%s]' on environment '%s' for stack '%s'\n", command, environment, service)
	fmt.Fprintf(os.Stdout, "TBD REMOVE----Task Definition Input %s\n", ecsRunTaskInput)
	log.Debugf("Executing command '[%s]' on environment '%s' for stack '%s'\n", command, environment, service)

	resp, err := taskMgr.ecsAPI.RunTask(ecsRunTaskInput)
	if err != nil {
		return Empty, err
	}

	fmt.Fprintf(os.Stdout, "TBD REMOVE----Response: %s\n", resp)
	log.Debugf("ECS Task Response: %s\n", resp)
	return resp.String(), nil
}
