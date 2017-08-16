package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strconv"
	"strings"
)

// NewServiceDeployer create a new workflow for deploying a service in an environment
func NewServiceDeployer(ctx *common.Context, environmentName string, tag string) Executor {

	workflow := new(serviceWorkflow)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = ctx.Config.Repo.Slug

	stackParams := make(map[string]string)

	return newPipelineExecutor(
		workflow.serviceLoader(ctx, tag, ""),
		workflow.serviceEnvironmentLoader(environmentName, ctx.StackManager),
		workflow.serviceApplyCommonParams(&ctx.Config.Service, stackParams, environmentName, ctx.StackManager, ctx.ElbManager, ctx.ParamManager),
		newConditionalExecutor(workflow.isEcsProvider(),
			newPipelineExecutor(
				workflow.serviceRepoUpserter(&ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceApplyEcsParams(&ctx.Config.Service, stackParams),
				workflow.serviceEcsDeployer(&ctx.Config.Service, stackParams, environmentName, ctx.StackManager, ctx.StackManager),
			),
			newPipelineExecutor(
				workflow.serviceBucketUpserter(&ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceAppUpserter(&ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceApplyEc2Params(stackParams),
				workflow.serviceEc2Deployer(&ctx.Config.Service, stackParams, environmentName, ctx.StackManager, ctx.StackManager),
			),
		),
	)
}

func getMaxPriority(elbRuleLister common.ElbRuleLister, listenerArn string) int {
	rules, err := elbRuleLister.ListRules(listenerArn)
	if err != nil {
		log.Debugf("Error finding max priority for listener '%s': %v", listenerArn, err)
		return 0
	}
	maxPriority := 0
	for _, rule := range rules {
		priority, _ := strconv.Atoi(common.StringValue(rule.Priority))
		if priority > maxPriority {
			maxPriority = priority
		}
	}
	return maxPriority
}

func (workflow *serviceWorkflow) serviceEnvironmentLoader(environmentName string, stackWaiter common.StackWaiter) Executor {
	return func() error {
		lbStackName := common.CreateStackName(common.StackTypeLoadBalancer, environmentName)
		workflow.lbStack = stackWaiter.AwaitFinalStatus(lbStackName)

		envStackName := common.CreateStackName(common.StackTypeEnv, environmentName)
		workflow.envStack = stackWaiter.AwaitFinalStatus(envStackName)

		if workflow.envStack == nil {
			return fmt.Errorf("Unable to find stack '%s' for environment '%s'", envStackName, environmentName)
		}

		if workflow.isEcsProvider()() {
			workflow.artifactProvider = common.ArtifactProviderEcr
		} else {
			workflow.artifactProvider = common.ArtifactProviderS3
		}

		return nil
	}
}

func (workflow *serviceWorkflow) serviceApplyEcsParams(service *common.Service, params map[string]string) Executor {
	return func() error {

		params["EcsCluster"] = fmt.Sprintf("%s-EcsCluster", workflow.envStack.Name)
		params["ImageUrl"] = workflow.serviceImage

		if service.CPU != 0 {
			params["ServiceCpu"] = strconv.Itoa(service.CPU)
		}
		if service.Memory != 0 {
			params["ServiceMemory"] = strconv.Itoa(service.Memory)
		}

		return nil
	}
}

func (workflow *serviceWorkflow) serviceApplyEc2Params(params map[string]string) Executor {
	return func() error {

		params["AppName"] = workflow.appName
		params["RevisionBucket"] = workflow.appRevisionBucket
		params["RevisionKey"] = workflow.appRevisionKey
		params["RevisionBundleType"] = "zip"

		for _, key := range [...]string{
			"SshAllow",
			"InstanceType",
			"ImageId",
			"MaxSize",
			"KeyName",
			"ScaleInThreshold",
			"ScaleOutThreshold",
			"HttpProxy",
			"ConsulServerAutoScalingGroup",
			"ElbSecurityGroup",
			"ConsulRpcClientSecurityGroup",
			"InstanceSecurityGroup",
		} {
			params[key] = workflow.envStack.Outputs[key]
		}

		for _, key := range [...]string{
			"InstanceSubnetIds",
		} {
			params[key] = workflow.envStack.Parameters[key]
		}

		return nil
	}
}

func (workflow *serviceWorkflow) serviceApplyCommonParams(service *common.Service, params map[string]string, environmentName string, stackWaiter common.StackWaiter, elbRuleLister common.ElbRuleLister, paramGetter common.ParamGetter) Executor {
	return func() error {
		params["VpcId"] = fmt.Sprintf("%s-VpcId", workflow.envStack.Name)

		nextAvailablePriority := 0
		if workflow.lbStack.Outputs["ElbHttpListenerArn"] != "" {
			params["ElbHttpListenerArn"] = fmt.Sprintf("%s-ElbHttpListenerArn", workflow.lbStack.Name)
			nextAvailablePriority = 1 + getMaxPriority(elbRuleLister, workflow.lbStack.Outputs["ElbHttpListenerArn"])
		}
		if workflow.lbStack.Outputs["ElbHttpsListenerArn"] != "" {
			params["ElbHttpsListenerArn"] = fmt.Sprintf("%s-ElbHttpsListenerArn", workflow.lbStack.Name)
			if nextAvailablePriority == 0 {
				nextAvailablePriority = 1 + getMaxPriority(elbRuleLister, workflow.lbStack.Outputs["ElbHttpsListenerArn"])
			}
		}

		dbStackName := common.CreateStackName(common.StackTypeDatabase, workflow.serviceName, environmentName)
		dbStack := stackWaiter.AwaitFinalStatus(dbStackName)
		if dbStack != nil {
			params["DatabaseName"] = dbStack.Outputs["DatabaseName"]
			params["DatabaseEndpointAddress"] = dbStack.Outputs["DatabaseEndpointAddress"]
			params["DatabaseEndpointPort"] = dbStack.Outputs["DatabaseEndpointPort"]
			params["DatabaseMasterUsername"] = dbStack.Outputs["DatabaseMasterUsername"]

			dbPass, _ := paramGetter.GetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"))
			params["DatabaseMasterPassword"] = dbPass
		}

		svcStackName := common.CreateStackName(common.StackTypeService, workflow.serviceName, environmentName)
		svcStack := stackWaiter.AwaitFinalStatus(svcStackName)

		if workflow.priority > 0 {
			params["PathListenerRulePriority"] = strconv.Itoa(workflow.priority)
			params["HostListenerRulePriority"] = strconv.Itoa(workflow.priority + 1)
		} else if svcStack != nil {
			// no value in config, and this is an update...use prior value
			params["PathListenerRulePriority"] = ""
			params["HostListenerRulePriority"] = ""
		} else {
			// no value in config, and this is a create...use next available
			params["PathListenerRulePriority"] = strconv.Itoa(nextAvailablePriority)
			params["HostListenerRulePriority"] = strconv.Itoa(nextAvailablePriority + 1)
		}

		params["ServiceName"] = workflow.serviceName
		if service.Port != 0 {
			params["ServicePort"] = strconv.Itoa(service.Port)
		}
		if service.Protocol != "" {
			params["ServiceProtocol"] = strings.ToUpper(service.Protocol)
		}
		if service.HealthEndpoint != "" {
			params["ServiceHealthEndpoint"] = service.HealthEndpoint
		}
		if service.DesiredCount != 0 {
			params["ServiceDesiredCount"] = strconv.Itoa(service.DesiredCount)
		}
		if len(service.PathPatterns) > 0 {
			params["PathPattern"] = strings.Join(service.PathPatterns, ",")
		}
		if len(service.HostPatterns) > 0 {
			params["HostPattern"] = strings.Join(service.HostPatterns, ",")
		}

		return nil
	}
}

func (workflow *serviceWorkflow) serviceEc2Deployer(service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {

		log.Noticef("Deploying service '%s' to '%s'", workflow.serviceName, environmentName)

		svcStackName := common.CreateStackName(common.StackTypeService, workflow.serviceName, environmentName)

		resolveServiceEnvironment(service, environmentName)
		overrides := common.GetStackOverrides(svcStackName)
		template, err := templates.NewTemplate("service-ec2.yml", service, overrides)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(svcStackName, template, stackParams, buildServiceTags(workflow.serviceName, environmentName, workflow.envStack.Outputs["provider"], common.StackTypeService, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", svcStackName)
		stack := stackWaiter.AwaitFinalStatus(svcStackName)
		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", svcStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}
}

func (workflow *serviceWorkflow) serviceEcsDeployer(service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Deploying service '%s' to '%s' from '%s'", workflow.serviceName, environmentName, workflow.serviceImage)

		svcStackName := common.CreateStackName(common.StackTypeService, workflow.serviceName, environmentName)

		resolveServiceEnvironment(service, environmentName)
		overrides := common.GetStackOverrides(svcStackName)
		template, err := templates.NewTemplate("service-ecs.yml", service, overrides)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(svcStackName, template, stackParams, buildServiceTags(workflow.serviceName, environmentName, workflow.envStack.Outputs["provider"], common.StackTypeService, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", svcStackName)
		stack := stackWaiter.AwaitFinalStatus(svcStackName)
		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", svcStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}

}

func resolveServiceEnvironment(service *common.Service, environment string) {
	for key, value := range service.Environment {
		switch value.(type) {
		case map[interface{}]interface{}:
			found := false
			for env, v := range value.(map[interface{}]interface{}) {
				if env.(string) == environment {
					service.Environment[key] = v.(string)
					found = true
				}
			}
			if found != true {
				service.Environment[key] = ""
			}
		case string:
			// do nothing
		default:
			log.Warningf("Unable to resolve environment '%s': %v", key, value)
		}

	}
}

func buildServiceTags(serviceName string, environmentName string, envProvider string, stackType common.StackType, codeRevision string, repoName string) map[string]string {
	return map[string]string{
		"type":        string(stackType),
		"environment": environmentName,
		"provider":    envProvider,
		"service":     serviceName,
		"revision":    codeRevision,
		"repo":        repoName,
	}
}
