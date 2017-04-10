package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"strings"
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

		stack := stackWaiter.AwaitFinalStatus(pipelineStackName)
		if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}
		return nil
	}
}
