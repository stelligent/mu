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

	ecsImportParams := make(map[string]string)

	return newWorkflow(
		workflow.databaseInput(ctx, ""),
		workflow.databaseEnvironmentLoader(environmentName, ctx.StackManager, ecsImportParams, ctx.ElbManager),
		workflow.databaseDeployer(&ctx.Config.Service, ecsImportParams, environmentName, ctx.StackManager, ctx.StackManager, ctx.RdsManager),
	)
}

func (workflow *databaseWorkflow) databaseEnvironmentLoader(environmentName string, stackWaiter common.StackWaiter, ecsImportParams map[string]string, elbRuleLister common.ElbRuleLister) Executor {
	return func() error {
		ecsStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		ecsStack := stackWaiter.AwaitFinalStatus(ecsStackName)

		if ecsStack == nil {
			return fmt.Errorf("Unable to find stack '%s' for environment '%s'", ecsStackName, environmentName)
		}

		ecsImportParams["VpcId"] = fmt.Sprintf("%s-VpcId", ecsStackName)
		ecsImportParams["EcsInstanceSecurityGroup"] = fmt.Sprintf("%s-EcsInstanceSecurityGroup", ecsStackName)
		ecsImportParams["EcsSubnetIds"] = fmt.Sprintf("%s-EcsSubnetIds", ecsStackName)

		return nil
	}
}

func (workflow *databaseWorkflow) databaseDeployer(service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter, rdsSetter common.RdsIamAuthenticationSetter) Executor {
	return func() error {

		if service.Database.Name == "" {
			log.Noticef("Skipping database since database.name is unset")
			return nil
		}

		log.Noticef("Deploying database '%s' to '%s'", workflow.serviceName, environmentName)

		dbStackName := common.CreateStackName(common.StackTypeDatabase, workflow.serviceName, environmentName)

		overrides := common.GetStackOverrides(dbStackName)
		template, err := templates.NewTemplate("database.yml", service, overrides)
		if err != nil {
			return err
		}

		stackParams["DatabaseName"] = service.Database.Name

		if service.Database.Engine != "" {
			stackParams["DatabaseEngine"] = service.Database.Engine
		}

		if service.Database.InstanceClass != "" {
			stackParams["DatabaseInstanceClass"] = service.Database.InstanceClass
		}
		if service.Database.AllocatedStorage != "" {
			stackParams["DatabaseStorage"] = service.Database.AllocatedStorage
		}
		if service.Database.MasterUsername != "" {
			stackParams["DatabaseMasterUsername"] = service.Database.MasterUsername
		}

		//DatabaseMasterPassword:
		stackParams["DatabaseMasterPassword"] = "changeme"

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

		// update IAM Authentication
		return rdsSetter.SetIamAuthentication(stack.Outputs["DatabaseIdentifier"], service.Database.IamAuthentication, service.Database.Engine)
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
