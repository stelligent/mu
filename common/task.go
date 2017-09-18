package common

import (
	"github.com/aws/aws-sdk-go/service/ecs"
)

// TaskContainerLister for listing tasks with containers
type TaskContainerLister interface {
	ListTasks(ctx *Context, environment string, serviceName string) ([]Task, error)
}

// TaskStopper for restarting tasks
type TaskStopper interface {
	StopTask(ctx *Context, environment string, task string) error
}

// ECSRunTaskResult describes the output result from ECS call to RunTask
type ECSRunTaskResult *ecs.RunTaskOutput

// TaskCommandExecutor for executing commands against an environment
type TaskCommandExecutor interface {
	ExecuteCommand(ctx *Context, task Task) (ECSRunTaskResult, error)
}

// TaskManager composite of all task capabilities
type TaskManager interface {
	TaskContainerLister
	TaskStopper
	TaskCommandExecutor
}
