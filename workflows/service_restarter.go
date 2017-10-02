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
		workflow.serviceRestarter(ctx.Config.Namespace, ctx.TaskManager, environmentName, serviceName, batchSize),
	)
}

func (workflow *serviceWorkflow) serviceRestarter(namespace string, taskManager common.TaskManager, environmentName string, serviceName string, batchSize int) Executor {
	return func() error {
		tasks, err := taskManager.ListTasks(namespace, environmentName, serviceName)

		if err != nil {
			return err
		}

		for taskIdx, task := range tasks {
			log.Noticef("Stopping task %s in environment %s", task.Name, environmentName)
			stopErr := taskManager.StopTask(namespace, environmentName, task.Name)
			if stopErr != nil {
				fmt.Println(stopErr)
			}

			// Polling for same length task lists
			if (taskIdx+1)%batchSize == 0 {

				newTaskList, _ := taskManager.ListTasks(namespace, environmentName, serviceName)
				for len(newTaskList) != len(tasks) {
					duration := time.Duration(PollDelay) * time.Second
					time.Sleep(duration)
					newTaskList, _ = taskManager.ListTasks(namespace, environmentName, serviceName)
				}
			}
		}

		return nil
	}
}
