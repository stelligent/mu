package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewPipelineTerminator create a new workflow for terminating a pipeline
func NewPipelineTerminator(ctx *common.Context, serviceName string) Executor {

	workflow := new(pipelineWorkflow)

	return newWorkflow(
		workflow.serviceFinder(serviceName, ctx),
		workflow.pipelineTerminator(ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *pipelineWorkflow) pipelineTerminator(stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Terminating Pipeline '%s' ...", workflow.serviceName)
		pipelineStackName := common.CreateStackName(common.StackTypePipeline, workflow.serviceName)
		err := stackDeleter.DeleteStack(pipelineStackName)
		if err != nil {
			return err
		}

		stackWaiter.AwaitFinalStatus(pipelineStackName)
		return nil
	}
}
