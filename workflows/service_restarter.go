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
		workflow.serviceRestarter(ctx.Config.Namespace, ctx.TaskManager, environmentName, batchSize),
	)
}

func (workflow *serviceWorkflow) serviceRestarter(namespace string, taskManager common.TaskManager, environmentName string, batchSize int) Executor {
	return func() error {
		tasks, err := taskManager.ListTasks(namespace, environmentName, workflow.serviceName)

		log.Noticef("Found %v tasks for service %s in environment %s", len(tasks), workflow.serviceName, environmentName)

		if err != nil {
			return err
		}

		for taskIdx, task := range tasks {
			log.Noticef("Restarting task %s in environment %s", task.Name, environmentName)
			stopErr := taskManager.StopTask(namespace, environmentName, task.Name)
			if stopErr != nil {
				fmt.Println(stopErr)
			}

			// Polling for same length task lists
			if (taskIdx+1)%batchSize == 0 {

				for countRunningTasks(namespace, taskManager, environmentName, workflow.serviceName) != len(tasks) {
					duration := time.Duration(PollDelay) * time.Second
					time.Sleep(duration)
				}
			}
		}

		return nil
	}
}

func countRunningTasks(namespace string, taskManager common.TaskManager, environmentName string, serviceName string) int {
	newTaskList, _ := taskManager.ListTasks(namespace, environmentName, serviceName)
	runningCount := 0
	for _, newTask := range newTaskList {
		if newTask.Status == "RUNNING" {
			runningCount++
		}
	}
	log.Debugf("Environment: %s, Service: %s, Running Tasks: %v", environmentName, serviceName, runningCount)
	return runningCount
}
