package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"strings"
)

// NewServiceUndeployer create a new workflow for undeploying a service in an environment
func NewServiceUndeployer(ctx *common.Context, serviceName string, environmentName string) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceInput(ctx, serviceName),
		workflow.serviceUndeployer(environmentName, ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *serviceWorkflow) serviceUndeployer(environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Undeploying service '%s' from '%s'", workflow.serviceName, environmentName)
		svcStackName := common.CreateStackName(common.StackTypeService, workflow.serviceName, environmentName)
		svcStack := stackWaiter.AwaitFinalStatus(svcStackName)
		if svcStack != nil {
			err := stackDeleter.DeleteStack(svcStackName)
			if err != nil {
				return err
			}
			svcStack = stackWaiter.AwaitFinalStatus(svcStackName)
			if svcStack != nil && !strings.HasSuffix(svcStack.Status, "_COMPLETE") {
				return fmt.Errorf("Ended in failed status %s %s", svcStack.Status, svcStack.StatusReason)
			}
		} else {
			log.Info("  Stack is alredy deleted.")
		}

		return nil
	}
}
