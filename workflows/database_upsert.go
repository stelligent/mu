package workflows

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/stelligent/mu/common"
)

// NewDatabaseUpserter create a new workflow for deploying a database in an environment
func NewDatabaseUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(databaseWorkflow)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = ctx.Config.Repo.Slug

	ecsImportParams := make(map[string]string)

	return newPipelineExecutor(
		workflow.databaseInput(ctx, "", environmentName),
		workflow.databaseEnvironmentLoader(ctx.Config.Namespace, environmentName, ctx.StackManager, ecsImportParams, ctx.ElbManager),
		workflow.databaseRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, environmentName),
		workflow.databaseDeployer(ctx.Config.Namespace, &ctx.Config.Service, ecsImportParams, environmentName, ctx.StackManager, ctx.StackManager, ctx.RdsManager, ctx.ParamManager),
	)
}

func (workflow *databaseWorkflow) databaseEnvironmentLoader(namespace string, environmentName string, stackWaiter common.StackWaiter, ecsImportParams map[string]string, elbRuleLister common.ElbRuleLister) Executor {
	return func() error {
		ecsStackName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
		ecsStack := stackWaiter.AwaitFinalStatus(ecsStackName)

		if ecsStack == nil {
			return fmt.Errorf("Unable to find stack '%s' for environment '%s'", ecsStackName, environmentName)
		}

		ecsImportParams["VpcId"] = fmt.Sprintf("%s-VpcId", ecsStackName)
		ecsImportParams["InstanceSecurityGroup"] = fmt.Sprintf("%s-InstanceSecurityGroup", ecsStackName)
		ecsImportParams["InstanceSubnetIds"] = fmt.Sprintf("%s-InstanceSubnetIds", ecsStackName)

		return nil
	}
}

func (workflow *databaseWorkflow) databaseRolesetUpserter(rolesetUpserter common.RolesetUpserter, rolesetGetter common.RolesetGetter, environmentName string) Executor {
	return func() error {

		err := rolesetUpserter.UpsertCommonRoleset()
		if err != nil {
			return err
		}

		commonRoleset, err := rolesetGetter.GetCommonRoleset()
		if err != nil {
			return err
		}

		workflow.cloudFormationRoleArn = commonRoleset["CloudFormationRoleArn"]

		err = rolesetUpserter.UpsertServiceRoleset(environmentName, workflow.serviceName, workflow.appRevisionBucket)
		if err != nil {
			return err
		}

		serviceRoleset, err := rolesetGetter.GetServiceRoleset(environmentName, workflow.serviceName)
		if err != nil {
			return err
		}
		workflow.databaseKeyArn = serviceRoleset["DatabaseKeyArn"]

		return nil
	}
}

func (workflow *databaseWorkflow) databaseDeployer(namespace string, service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter, rdsSetter common.RdsIamAuthenticationSetter, paramManager common.ParamManager) Executor {
	return func() error {

		if service.Database.Name == "" {
			log.Noticef("Skipping database since database.name is unset")
			return nil
		}

		log.Noticef("Deploying database '%s' to '%s'", workflow.serviceName, environmentName)

		dbStackName := common.CreateStackName(namespace, common.StackTypeDatabase, workflow.serviceName, environmentName)

		stackParams["DatabaseName"] = service.Database.Name

		common.NewMapElementIfNotEmpty(stackParams, "DatabaseEngine", service.Database.Engine)
		common.NewMapElementIfNotEmpty(stackParams, "DatabaseInstanceClass", service.Database.InstanceClass)
		common.NewMapElementIfNotEmpty(stackParams, "DatabaseStorage", service.Database.AllocatedStorage)

			stackParams["DatabaseMasterUsername"] = "admin"
		common.NewMapElementIfNotEmpty(stackParams, "DatabaseMasterUsername", service.Database.MasterUsername)

		//DatabaseMasterPassword:
		dbPass, err := paramManager.GetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"))
		if err != nil {
			log.Warningf("Error with GetParam for DatabaseMasterPassword, assuming empty: %s", err)
			dbPass = ""
		}
		if dbPass == "" {
			dbPass = randomPassword(32)
			err := paramManager.SetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"), dbPass, workflow.databaseKeyArn)
			if err != nil {
				return err
			}
		}
		stackParams["DatabaseMasterPassword"] = dbPass

		stackParams["DatabaseKeyArn"] = workflow.databaseKeyArn

		tags := createTagMap(&DatabaseTags{
			Environment: environmentName,
			Type:        common.StackTypeDatabase,
			Service:     workflow.serviceName,
			Revision:    workflow.codeRevision,
			Repo:        workflow.repoName,
		})

		err = stackUpserter.UpsertStack(dbStackName, "database.yml", service, stackParams, tags, workflow.cloudFormationRoleArn)

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
		if stack.Outputs["DatabaseIdentifier"] != "" {
			return rdsSetter.SetIamAuthentication(stack.Outputs["DatabaseIdentifier"], service.Database.IamAuthentication, service.Database.Engine)
		}

		return nil
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
