package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
)

// DatabaseSetPassword sets a database password for an environment
func DatabaseSetPassword(ctx *common.Context, environmentName string, serviceNameIn string, newPassword string) Executor {
	return func() error {
		workflow := new(databaseWorkflow)
		workflow.codeRevision = ctx.Config.Repo.Revision
		workflow.repoName = ctx.Config.Repo.Slug
		serviceName := serviceNameIn
		if serviceName == "" {
			serviceName = workflow.serviceName
		}

		dbStackName := common.CreateStackName(common.StackTypeDatabase, serviceName, environmentName)
		log.Noticef("Storing password for dbStackName:%s", dbStackName)
		if err := ctx.ParamManager.SetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"), newPassword); err != nil {
			return err
		}
		return nil
	}
}

// DatabaseGetPassword gets a database password for an environment
func DatabaseGetPassword(ctx *common.Context, environmentName string, serviceNameIn string) Executor {
	return func() error {
		workflow := new(databaseWorkflow)
		workflow.codeRevision = ctx.Config.Repo.Revision
		workflow.repoName = ctx.Config.Repo.Slug
		fmt.Printf("ServiceName is %s\n", serviceNameIn)
		serviceName := serviceNameIn
		if serviceName == "" {
			serviceName = workflow.serviceName
		}
		fmt.Printf("ServiceName is %s\n", serviceName)

		dbStackName := common.CreateStackName(common.StackTypeDatabase, serviceName, environmentName)
		dbPass, _ := ctx.ParamManager.GetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"))
		log.Noticef("%s", dbPass)
		return nil
	}
}
