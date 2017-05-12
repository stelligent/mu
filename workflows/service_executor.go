package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewServiceExecutor create a new workflow for executing a command in an environment
func NewServiceExecutor(ctx *common.Context, task common.Task) Executor {

	workflow := new(environmentWorkflow)
	if len(task.Service) == common.Zero {
		task.Service = ctx.Config.Service.Name
	}

	return newWorkflow(
		workflow.serviceTaskExecutor(ctx.TaskManager, task),
	)
}

func newServiceExecutor(taskManager common.TaskManager, task common.Task) Executor {
	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.serviceTaskExecutor(taskManager, task),
	)
}

func (workflow *environmentWorkflow) serviceTaskExecutor(taskManager common.TaskManager, task common.Task) Executor {
	return func() error {
		log.Notice(common.SvcCmdTaskExecutingLog)
		result, err := taskManager.ExecuteCommand(task)
		if err != nil {
			log.Noticef(common.SvcCmdTaskErrorLog, err)
			return err
		}
		log.Noticef(common.SvcCmdTaskResultLog, result)
		return nil
	}
}
