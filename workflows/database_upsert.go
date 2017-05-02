package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
)

// NewDatabaseUpserter create a new workflow for deploying a database in an environment
func NewDatabaseUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(databaseWorkflow)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = fmt.Sprintf("%s/%s", ctx.Config.Repo.OrgName, ctx.Config.Repo.Name)

	svcWorkflow := new(serviceWorkflow)

	ecsImportParams := make(map[string]string)

	return newWorkflow(
		workflow.databaseInput(ctx, ""),
		svcWorkflow.serviceLoader(ctx, ""),
		svcWorkflow.serviceEnvironmentLoader(environmentName, ctx.StackManager, ecsImportParams, ctx.ElbManager),
		workflow.databaseDeployer(&ctx.Config.Service, ecsImportParams, environmentName, ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *databaseWorkflow) databaseDeployer(service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Deploying database '%s' to '%s'", workflow.serviceName, environmentName)

		stackParams["DatabaseName"] = workflow.serviceName

		dbStackName := common.CreateStackName(common.StackTypeDatabase, workflow.serviceName, environmentName)

		overrides := common.GetStackOverrides(dbStackName)
		template, err := templates.NewTemplate("database.yml", service, overrides)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(dbStackName, template, stackParams, buildDatabaseTags(workflow.serviceName, environmentName, common.StackTypeDatabase, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", dbStackName)
		stack := stackWaiter.AwaitFinalStatus(dbStackName)
		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", dbStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}
}
func buildDatabaseTags(serviceName string, environmentName string, stackType common.StackType, codeRevision string, repoName string) map[string]string {
	return map[string]string{
		"type":        string(stackType),
		"environment": environmentName,
		"service":     serviceName,
		"revision":    codeRevision,
		"repo":        repoName,
	}
}
