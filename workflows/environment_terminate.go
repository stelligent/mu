package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewEnvironmentTerminator create a new workflow for terminating an environment
func NewEnvironmentTerminator(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentEcsTerminator(environmentName, ctx.StackManager, ctx.StackManager),
		workflow.environmentVpcTerminator(environmentName, ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *environmentWorkflow) environmentEcsTerminator(environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		envStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		err := stackDeleter.DeleteStack(envStackName)
		if err != nil {
			return err
		}

		stackWaiter.AwaitFinalStatus(envStackName)
		return nil
	}
}
func (workflow *environmentWorkflow) environmentVpcTerminator(environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		vpcStackName := common.CreateStackName(common.StackTypeVpc, environmentName)
		err := stackDeleter.DeleteStack(vpcStackName)
		if err != nil {
			log.Debugf("Unable to delete VPC, but ignoring error: %v", err)
		}

		stackWaiter.AwaitFinalStatus(vpcStackName)
		return nil
	}
}
