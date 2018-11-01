package common

import (
	"github.com/aws/aws-sdk-go/service/batch"
)

/*
// TaskContainerLister for listing tasks with containers
type TaskContainerLister interface {
	ListTasks(namespace string, environment string, serviceName string) ([]Task, error)
}

// TaskStopper for restarting tasks
type TaskStopper interface {
	StopTask(namespace string, environment string, task string) error
}

*/

// BatchJobSubmitResult describes the output result from Batch call to SubmitJob
type BatchJobSubmitResult *batch.SubmitJobOutput

// BatchJobExecutor for executing batch jobs against an environment
type BatchJobExecutor interface {
	ExecuteCommand(namespace string, task Task) (BatchJobSubmitResult, error)
}

// BatchManager composite of all task capabilities
type BatchManager interface {
	//TaskContainerLister
	//TaskStopper
	BatchJobExecutor
}
