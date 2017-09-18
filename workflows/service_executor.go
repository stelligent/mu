package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewServiceExecutor create a new workflow for executing a command in an environment
func NewServiceExecutor(ctx *common.Context, task common.Task) Executor {

	workflow := new(environmentWorkflow)
	if len(task.Service) == Zero {
		task.Service = ctx.Config.Service.Name
	}

	return newPipelineExecutor(
		workflow.serviceTaskExecutor(ctx, ctx.TaskManager, task),
	)
}

func newServiceExecutor(ctx *common.Context, taskManager common.TaskManager, task common.Task) Executor {
	workflow := new(environmentWorkflow)

	return newPipelineExecutor(
		workflow.serviceTaskExecutor(ctx, taskManager, task),
	)
}

func (workflow *environmentWorkflow) serviceTaskExecutor(ctx *common.Context, taskManager common.TaskManager, task common.Task) Executor {
	return func() error {
		log.Notice(SvcCmdTaskExecutingLog)
		result, err := taskManager.ExecuteCommand(ctx, task)
		if err != nil {
			log.Noticef(SvcCmdTaskErrorLog, err)
			return err
		}
		log.Noticef(SvcCmdTaskResultLog, result)
		return nil
	}
}
