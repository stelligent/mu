package workflows

import (
	"fmt"
	"strings"

	"github.com/stelligent/mu/common"
)

// NewDatabaseTerminator create a new workflow for terminating a database in an environment
func NewDatabaseTerminator(ctx *common.Context, serviceName string, environmentName string) Executor {

	workflow := new(databaseWorkflow)

	return newPipelineExecutor(
		workflow.databaseInput(ctx, serviceName, environmentName),
		workflow.databaseTerminator(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager, ctx.ParamManager),
	)
}

func (workflow *databaseWorkflow) databaseTerminator(namespace string, environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter, paramDeleter common.ParamDeleter) Executor {
	return func() error {
		log.Noticef("Deleting database '%s' from '%s'", workflow.serviceName, environmentName)
		dbStackName := common.CreateStackName(namespace, common.StackTypeDatabase, workflow.serviceName, environmentName)
		dbStack := stackWaiter.AwaitFinalStatus(dbStackName)
		if dbStack != nil {
			err := stackDeleter.DeleteStack(dbStackName)
			if err != nil {
				return err
			}
			dbStack = stackWaiter.AwaitFinalStatus(dbStackName)
			if dbStack != nil && !strings.HasSuffix(dbStack.Status, "_COMPLETE") {
				return fmt.Errorf("Ended in failed status %s %s", dbStack.Status, dbStack.StatusReason)
			}
		} else {
			log.Info("  Stack is already deleted.")
		}

		paramDeleter.DeleteParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"))
		return nil
	}
}
