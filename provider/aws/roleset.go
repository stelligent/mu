package aws

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/stelligent/mu/common"
)

type iamRolesetManager struct {
	context *common.Context
}

func newRolesetManager(ctx *common.Context) (common.RolesetManager, error) {
	return &iamRolesetManager{
		context: ctx,
	}, nil
}

func (rolesetMgr *iamRolesetManager) getRolesetFromStack(names ...string) common.Roleset {
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, names...)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)

	if stack == nil {
		return make(map[string]string)
	}

	return stack.Outputs
}

func overrideRole(roleset common.Roleset, roleType string, roleArn string) {
	if roleArn != "" {
		roleset[roleType] = roleArn
	}
}

func (rolesetMgr *iamRolesetManager) GetCommonRoleset() (common.Roleset, error) {
	roleset := rolesetMgr.getRolesetFromStack("common")

	overrideRole(roleset, "CloudFormationRoleArn", rolesetMgr.context.Config.Roles.CloudFormation)

	return roleset, nil
}

func (rolesetMgr *iamRolesetManager) GetEnvironmentRoleset(environmentName string) (common.Roleset, error) {
	roleset := rolesetMgr.getRolesetFromStack("environment", environmentName)

	for _, e := range rolesetMgr.context.Config.Environments {
		if strings.EqualFold(e.Name, environmentName) {
			overrideRole(roleset, "EC2InstanceProfileArn", e.Roles.EcsInstance)
			overrideRole(roleset, "ConsulClientTaskRoleArn", e.Roles.ConsulClientTask)
			overrideRole(roleset, "ConsulEC2InstanceProfileArn", e.Roles.ConsulInstance)
			overrideRole(roleset, "ConsulServerTaskRoleArn", e.Roles.ConsulServerTask)
			break
		}
	}

	return roleset, nil
}

func (rolesetMgr *iamRolesetManager) GetServiceRoleset(environmentName string, serviceName string) (common.Roleset, error) {
	roleset := rolesetMgr.getRolesetFromStack("service", serviceName, environmentName)

	overrideRole(roleset, "EC2InstanceProfileArn", rolesetMgr.context.Config.Service.Roles.Ec2Instance)
	overrideRole(roleset, "CodeDeployRoleArn", rolesetMgr.context.Config.Service.Roles.CodeDeploy)
	overrideRole(roleset, "EcsServiceRoleArn", rolesetMgr.context.Config.Service.Roles.EcsService)
	overrideRole(roleset, "EcsTaskRoleArn", rolesetMgr.context.Config.Service.Roles.EcsTask)

	return roleset, nil
}

func (rolesetMgr *iamRolesetManager) GetPipelineRoleset(serviceName string) (common.Roleset, error) {
	roleset := rolesetMgr.getRolesetFromStack("pipeline", serviceName)

	overrideRole(roleset, "CodePipelineRoleArn", rolesetMgr.context.Config.Service.Pipeline.Roles.Pipeline)
	overrideRole(roleset, "CodeBuildCIRoleArn", rolesetMgr.context.Config.Service.Pipeline.Roles.Build)
	overrideRole(roleset, "CodeBuildCDAcptRoleArn", rolesetMgr.context.Config.Service.Pipeline.Acceptance.Roles.CodeBuild)
	overrideRole(roleset, "CodeBuildCDProdRoleArn", rolesetMgr.context.Config.Service.Pipeline.Production.Roles.CodeBuild)
	overrideRole(roleset, "MuAcptRoleArn", rolesetMgr.context.Config.Service.Pipeline.Acceptance.Roles.Mu)
	overrideRole(roleset, "MuProdRoleArn", rolesetMgr.context.Config.Service.Pipeline.Production.Roles.Mu)

	return roleset, nil
}

func (rolesetMgr *iamRolesetManager) UpsertCommonRoleset() error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping upsert of common IAM roles.")
		return nil
	}
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "common")
	stackTags := map[string]string{
		"mu:type": "iam",
	}

	stackParams := map[string]string{
		"Namespace": rolesetMgr.context.Config.Namespace,
	}

	err := rolesetMgr.context.StackManager.UpsertStack(stackName, "common-iam.yml", nil, stackParams, stackTags, "")
	if err != nil {
		// ignore error if stack is in progress already
		if !strings.Contains(err.Error(), "_IN_PROGRESS state and can not be updated") {
			return err
		}
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack == nil {
		return fmt.Errorf("Unable to create stack %s", stackName)
	}
	if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}

	return nil
}

func (rolesetMgr *iamRolesetManager) UpsertEnvironmentRoleset(environmentName string) error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping upsert of environment IAM roles.")
		return nil
	}

	var environment *common.Environment
	for _, e := range rolesetMgr.context.Config.Environments {
		if strings.EqualFold(e.Name, environmentName) {
			if e.Provider == "" {
				e.Provider = common.EnvProviderEcs
			}
			environment = &e
			break
		}
	}
	if environment == nil {
		log.Warningf("unable to find environment named '%s' in configuration...skipping IAM roles", environmentName)
		return nil
	}

	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "environment", environmentName)
	stackTags := map[string]string{
		"mu:type":        "iam",
		"mu:environment": environmentName,
		"mu:provider":    string(environment.Provider),
		"mu:revision":    rolesetMgr.context.Config.Repo.Revision,
		"mu:repo":        rolesetMgr.context.Config.Repo.Name,
	}

	stackParams := map[string]string{
		"Namespace":       rolesetMgr.context.Config.Namespace,
		"EnvironmentName": environmentName,
	}

	if strings.EqualFold(environment.Discovery.Provider, "consul") {
		stackParams["EnableConsul"] = "true"
	}

	err := rolesetMgr.context.StackManager.UpsertStack(stackName, "env-iam.yml", environment, stackParams, stackTags, "")
	if err != nil {
		return err
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack == nil {
		return fmt.Errorf("Unable to create stack %s", stackName)
	}
	if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}

	return nil
}

func (rolesetMgr *iamRolesetManager) UpsertServiceRoleset(environmentName string, serviceName string) error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping upsert of service IAM roles.")
		return nil
	}
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "service", serviceName, environmentName)
	envProvider := ""
	for _, e := range rolesetMgr.context.Config.Environments {
		if strings.EqualFold(e.Name, environmentName) {
			if e.Provider == "" {
				envProvider = string(common.EnvProviderEcs)
			} else {
				envProvider = string(e.Provider)
			}
			break
		}
	}
	if envProvider == "" {
		log.Debugf("unable to find environment named '%s' in configuration...checking for existing stack", environmentName)
		envStackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeEnv, environmentName)
		envStack := rolesetMgr.context.StackManager.AwaitFinalStatus(envStackName)
		if envStack == nil {
			return fmt.Errorf("unable to find environment stack named '%s'", envStackName)
		}
		envProvider = envStack.Tags["provider"]
	}

	stackTags := map[string]string{
		"mu:type":        "iam",
		"mu:environment": environmentName,
		"mu:provider":    envProvider,
		"mu:service":     serviceName,
		"mu:revision":    rolesetMgr.context.Config.Repo.Revision,
		"mu:repo":        rolesetMgr.context.Config.Repo.Name,
	}

	stackParams := map[string]string{
		"Namespace":       rolesetMgr.context.Config.Namespace,
		"EnvironmentName": environmentName,
		"ServiceName":     serviceName,
		"Provider":        envProvider,
	}

	err := rolesetMgr.context.StackManager.UpsertStack(stackName, "service-iam.yml", rolesetMgr.context.Config.Service, stackParams, stackTags, "")
	if err != nil {
		return err
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack == nil {
		return fmt.Errorf("Unable to create stack %s", stackName)
	}
	if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}

	return nil
}

func (rolesetMgr *iamRolesetManager) UpsertPipelineRoleset(serviceName string) error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping upsert of pipeline IAM roles.")
		return nil
	}
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "pipeline", serviceName)
	stackTags := map[string]string{
		"mu:type":     "iam",
		"mu:service":  serviceName,
		"mu:revision": rolesetMgr.context.Config.Repo.Revision,
		"mu:repo":     rolesetMgr.context.Config.Repo.Name,
	}

	pipelineConfig := rolesetMgr.context.Config.Service.Pipeline

	stackParams := map[string]string{
		"Namespace":      rolesetMgr.context.Config.Namespace,
		"ServiceName":    serviceName,
		"SourceProvider": pipelineConfig.Source.Provider,
		"SourceRepo":     pipelineConfig.Source.Repo,
	}

	if pipelineConfig.Source.Provider == "S3" {
		repoParts := strings.Split(pipelineConfig.Source.Repo, "/")
		stackParams["SourceBucket"] = repoParts[0]
		stackParams["SourceObjectKey"] = strings.Join(repoParts[1:], "/")
	}

	if pipelineConfig.Acceptance.Environment != "" {
		stackParams["AcptEnv"] = pipelineConfig.Acceptance.Environment
	}

	if pipelineConfig.Production.Environment != "" {
		stackParams["ProdEnv"] = pipelineConfig.Production.Environment
	}

	stackParams["EnableBuildStage"] = strconv.FormatBool(!pipelineConfig.Build.Disabled)
	stackParams["EnableAcptStage"] = strconv.FormatBool(!pipelineConfig.Acceptance.Disabled)
	stackParams["EnableProdStage"] = strconv.FormatBool(!pipelineConfig.Production.Disabled)

	commonRoleset, err := rolesetMgr.GetCommonRoleset()
	if err != nil {
		return err
	}
	stackParams["AcptCloudFormationRoleArn"] = commonRoleset["CloudFormationRoleArn"]
	stackParams["ProdCloudFormationRoleArn"] = commonRoleset["CloudFormationRoleArn"]

	err = rolesetMgr.context.StackManager.UpsertStack(stackName, "pipeline-iam.yml", rolesetMgr.context.Config.Service.Pipeline, stackParams, stackTags, "")
	if err != nil {
		return err
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack == nil {
		return fmt.Errorf("Unable to create stack %s", stackName)
	}
	if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}

	return nil
}

func (rolesetMgr *iamRolesetManager) DeleteCommonRoleset() error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping delete of common IAM roles.")
		return nil
	}
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "common")
	err := rolesetMgr.context.StackManager.DeleteStack(stackName)
	if err != nil {
		return err
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}
	return nil
}

func (rolesetMgr *iamRolesetManager) DeleteEnvironmentRoleset(environmentName string) error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping delete of environment IAM roles.")
		return nil
	}
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "environment", environmentName)
	err := rolesetMgr.context.StackManager.DeleteStack(stackName)
	if err != nil {
		return err
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}
	return nil
}

func (rolesetMgr *iamRolesetManager) DeleteServiceRoleset(environmentName string, serviceName string) error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping delete of service IAM roles.")
		return nil
	}
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "service", serviceName, environmentName)
	err := rolesetMgr.context.StackManager.DeleteStack(stackName)
	if err != nil {
		return err
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}
	return nil
}

func (rolesetMgr *iamRolesetManager) DeletePipelineRoleset(serviceName string) error {
	if rolesetMgr.context.Config.DisableIAM {
		log.Infof("Skipping delete of pipeline IAM roles.")
		return nil
	}
	stackName := common.CreateStackName(rolesetMgr.context.Config.Namespace, common.StackTypeIam, "pipeline", serviceName)
	err := rolesetMgr.context.StackManager.DeleteStack(stackName)
	if err != nil {
		return err
	}

	log.Debugf("Waiting for stack '%s' to complete", stackName)
	stack := rolesetMgr.context.StackManager.AwaitFinalStatus(stackName)
	if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
		return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
	}
	return nil
}
