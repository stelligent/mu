package common

import (
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TaskContainerLister for listing tasks with containers
type TaskContainerLister interface {
	ListTasks(environment string, serviceName string) ([]Task, error)
}

// TaskRestarter for restarting tasks
type TaskRestarter interface {
	StopTask(environment string, task string) error
}

// ECSRunTaskResult describes the output result from ECS call to RunTask
type ECSRunTaskResult *ecs.RunTaskOutput

// TaskCommandExecutor for executing commands against an environment
type TaskCommandExecutor interface {
	ExecuteCommand(task Task) (ECSRunTaskResult, error)
}

// TaskManager composite of all task capabilities
type TaskManager interface {
	TaskContainerLister
	TaskRestarter
	TaskCommandExecutor
}
