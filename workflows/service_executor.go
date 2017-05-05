package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewServiceExecutor create a new workflow for executing a command in an environment
func NewServiceExecutor(ctx *common.Context, environmentName string, service string, command string) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.serviceTaskExecutor(environmentName, service, command, ctx.TaskManager),
	)
}

func (workflow *environmentWorkflow) serviceTaskExecutor(environmentName string, service string, command string, taskManager common.TaskManager) Executor {
	return func() error {
		log.Noticef(common.EnvCmdTaskExecutingLog, command, environmentName)
		result, err := taskManager.ExecuteCommand(environmentName, service, command)
		if err != nil {
			log.Noticef(common.EnvCmdTaskErrorLog, err)
			return err
		}
		log.Noticef(common.EnvCmdTaskResultLog, command, environmentName, result)
		return nil
	}
}
