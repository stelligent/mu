package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewEnvironmentExecutor create a new workflow for executing a command in an environment
func NewEnvironmentExecutor(ctx *common.Context, environmentName string, command string) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentTaskExecutor(environmentName, ctx.StackManager, ctx.TaskManager, command),
	)
}

func (workflow *environmentWorkflow) environmentTaskExecutor(environmentName string, stackManager common.StackManager,
	taskManager common.TaskManager, command string) Executor {
	return func() error {
		log.Noticef(common.EnvCmdTaskExecutingLog, command, environmentName)

		result, err := taskManager.ExecuteCommand(environmentName, command)
		if err != nil {
			log.Noticef(common.EnvCmdTaskErrorLog, err)
			return err
		}
		log.Noticef(common.EnvCmdTaskResultLog, command, environmentName, result)
		return nil
	}
}
