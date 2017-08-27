package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewServiceRestarter create a new workflow for a rolling restart
func NewServiceRestarter(ctx *common.Context, serviceName string, batchSize int) Executor {
	workflow := new(serviceWorkflow)

	return newPipelineExecutor(
		workflow.serviceInput(ctx, serviceName),
	)
}