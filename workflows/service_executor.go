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
		workflow.serviceTaskExecutor(ctx.Config.Namespace, ctx.TaskManager, task),
	)
}

func newServiceExecutor(namespace string, taskManager common.TaskManager, task common.Task) Executor {
	workflow := new(environmentWorkflow)

	return newPipelineExecutor(
		workflow.serviceTaskExecutor(namespace, taskManager, task),
	)
}

func (workflow *environmentWorkflow) serviceTaskExecutor(namespace string, taskManager common.TaskManager, task common.Task) Executor {
	return func() error {
		log.Notice(SvcCmdTaskExecutingLog)
		result, err := taskManager.ExecuteCommand(namespace, task)
		if err != nil {
			log.Noticef(SvcCmdTaskErrorLog, err)
			return err
		}
		log.Noticef(SvcCmdTaskResultLog, result)
		return nil
	}
}
