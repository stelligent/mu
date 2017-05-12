package workflows

import (
	"crypto/rand"
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
)

// NewDatabaseUpserter create a new workflow for deploying a database in an environment
func NewDatabaseUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(databaseWorkflow)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = ctx.Config.Repo.Slug

	ecsImportParams := make(map[string]string)

	return newWorkflow(
		workflow.databaseInput(ctx, ""),
		workflow.databaseEnvironmentLoader(environmentName, ctx.StackManager, ecsImportParams, ctx.ElbManager),
		workflow.databaseDeployer(&ctx.Config.Service, ecsImportParams, environmentName, ctx.StackManager, ctx.StackManager, ctx.RdsManager, ctx.ParamManager),
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

func (workflow *databaseWorkflow) databaseDeployer(service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter, rdsSetter common.RdsIamAuthenticationSetter, paramManager common.ParamManager) Executor {
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
		dbPass, _ := paramManager.GetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"))
		if dbPass == "" {
			dbPass = randomPassword(32)
			err = paramManager.SetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"), dbPass)
			if err != nil {
				return err
			}
		}
		stackParams["DatabaseMasterPassword"] = dbPass

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

func randomPassword(length int) string {
	availableCharBytes := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	// Compute bitMask
	availableCharLength := len(availableCharBytes)
	if availableCharLength == 0 || availableCharLength > 256 {
		panic("availableCharBytes length must be greater than 0 and less than or equal to 256")
	}
	var bitLength byte
	var bitMask byte
	for bits := availableCharLength - 1; bits != 0; {
		bits = bits >> 1
		bitLength++
	}
	bitMask = 1<<bitLength - 1

	// Compute bufferSize
	bufferSize := length + length/3

	// Create random string
	result := make([]byte, length)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = make([]byte, length)
			_, err := rand.Read(randomBytes)
			if err != nil {
				log.Fatal("Unable to generate random bytes")
			}
		}
		// Mask bytes to get an index into the character slice
		if idx := int(randomBytes[j%length] & bitMask); idx < availableCharLength {
			result[i] = availableCharBytes[idx]
			i++
		}
	}

	return string(result)
}
