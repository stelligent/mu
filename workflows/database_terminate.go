package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"strings"
)

// NewDatabaseTerminator create a new workflow for terminating a database in an environment
func NewDatabaseTerminator(ctx *common.Context, serviceName string, environmentName string) Executor {

	workflow := new(databaseWorkflow)

	return newWorkflow(
		workflow.databaseInput(ctx, serviceName),
		workflow.databaseTerminator(environmentName, ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *databaseWorkflow) databaseTerminator(environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Undeploying service '%s' from '%s'", workflow.serviceName, environmentName)
		svcStackName := common.CreateStackName(common.StackTypeDatabase, workflow.serviceName, environmentName)
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
			log.Info("  Stack is already deleted.")
		}

		return nil
	}
}
