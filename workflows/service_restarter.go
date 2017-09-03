package workflows

import (
	"fmt"
	"time"

	"github.com/stelligent/mu/common"
)

// NewServiceRestarter create a new workflow for a rolling restart
func NewServiceRestarter(ctx *common.Context, environmentName string, serviceName string, batchSize int) Executor {

	workflow := new(serviceWorkflow)

	return newPipelineExecutor(
		workflow.serviceInput(ctx, serviceName),
		workflow.serviceRestarter(ctx.TaskManager, environmentName, serviceName, batchSize),
	)
}

func (workflow *serviceWorkflow) serviceRestarter(taskManager common.TaskManager, environmentName string, serviceName string, batchSize int) Executor {
	return func() error {
		tasks, err := taskManager.ListTasks(environmentName, serviceName)

		if err != nil {
			return err
		}

		idx := 0

		for idx < len(tasks) {
			stopErr := taskManager.StopTask(environmentName, tasks[idx].Name)
			if stopErr != nil {
				fmt.Println(stopErr)
			}

			// Polling for same length task lists
			newTaskList, _ := taskManager.ListTasks(environmentName, serviceName)
			for len(newTaskList) != len(tasks) {
				duration := time.Duration(PollDelay) * time.Second
				time.Sleep(duration)
				newTaskList, _ = taskManager.ListTasks(environmentName, serviceName)
			}
			idx++
		}

		return nil
	}
}
