package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"strings"
)

// NewPipelineTerminator create a new workflow for terminating a pipeline
func NewPipelineTerminator(ctx *common.Context, serviceName string) Executor {

	workflow := new(pipelineWorkflow)

	return newPipelineExecutor(
		workflow.serviceFinder(serviceName, ctx),
		workflow.pipelineTerminator(ctx.Config.Namespace, ctx.StackManager, ctx.StackManager),
		workflow.pipelineRolesetTerminator(ctx.RolesetManager),
	)
}

func (workflow *pipelineWorkflow) pipelineRolesetTerminator(rolesetDeleter common.RolesetDeleter) Executor {
	return func() error {
		err := rolesetDeleter.DeletePipelineRoleset(workflow.serviceName)
		if err != nil {
			return err
		}
		return nil
	}
}

func (workflow *pipelineWorkflow) pipelineTerminator(namespace string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Terminating Pipeline '%s' ...", workflow.serviceName)
		pipelineStackName := common.CreateStackName(namespace, common.StackTypePipeline, workflow.serviceName)
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
