package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stelligent/mu2/common"
	"os"
	"strings"
)

// TaskCommandExecutor for executing commands against an environment
type TaskCommandExecutor interface {
	ExecuteCommand(environmentName string, command string) (string, error)
}

// TaskManager composite of all task capabilities
type TaskManager interface {
	TaskCommandExecutor
}

type ecsTaskManager struct {
	dryrun bool
	cfnAPI cloudformationiface.CloudFormationAPI
	ecsAPI ecsiface.ECSAPI
}

func newTaskManager(sess *session.Session) (TaskManager, error) {
	log.Debug("Connecting to ECS service")
	ecsAPI := ecs.New(sess)
	log.Debug("Connecting to CloudFormation service")
	cfnAPI := cloudformation.New(sess)

	return &ecsTaskManager{
		dryrun: false,
		cfnAPI: cfnAPI,
		ecsAPI: ecsAPI,
	}, nil
}

// ExecuteCommand runs a command for a specific environment
func (taskMgr *ecsTaskManager) ExecuteCommand(environmentName string, command string) (string, error) {
	ecsStackName := common.CreateStackName(common.StackTypeCluster, environmentName)

	fmt.Fprintf(os.Stdout, "TBD REMOVE----Executing command [%s] on environment %s for stack %s\n", strings.TrimSpace(command), environmentName, ecsStackName)
	log.Debugf("Executing command '[%s]' on environment '%s' for stack '%s'\n", environmentName, command, ecsStackName)

	return command, nil
}
