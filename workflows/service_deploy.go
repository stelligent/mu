package workflows

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mholt/archiver"
	"github.com/stelligent/mu/common"
)

// NewServiceDeployer create a new workflow for deploying a service in an environment
func NewServiceDeployer(ctx *common.Context, environmentName string, tag string) Executor {

	workflow := new(serviceWorkflow)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = ctx.Config.Repo.Slug

	stackParams := make(map[string]string)

	return newPipelineExecutor(
		workflow.serviceLoader(ctx, tag, ""),
		workflow.serviceEnvironmentLoader(ctx.Config.Namespace, environmentName, ctx.StackManager),
		workflow.serviceApplyCommonParams(ctx.Config.Namespace, &ctx.Config.Service, stackParams, environmentName, ctx.StackManager, ctx.ElbManager, ctx.ParamManager),
		newConditionalExecutor(workflow.isEcsProvider(),
			newPipelineExecutor(
				workflow.serviceRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, environmentName),
				workflow.serviceRepoUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceApplyEcsParams(&ctx.Config.Service, stackParams, ctx.RolesetManager),
				workflow.serviceEcsDeployer(ctx.Config.Namespace, &ctx.Config.Service, stackParams, environmentName, ctx.StackManager, ctx.StackManager),
				workflow.serviceCreateSchedules(ctx.Config.Namespace, &ctx.Config.Service, environmentName, ctx.StackManager, ctx.StackManager),
			),
			newPipelineExecutor(
				workflow.serviceBucketUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, environmentName),
				workflow.serviceAppUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceApplyEc2Params(stackParams, ctx.RolesetManager),
				workflow.serviceEc2Deployer(ctx.Config.Namespace, &ctx.Config.Service, stackParams, environmentName, ctx.StackManager, ctx.StackManager),
				// TODO - placeholder for doing serviceCreateSchedules for EC2, leaving out-of-scope per @cplee
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

func (workflow *serviceWorkflow) serviceEnvironmentLoader(namespace string, environmentName string, stackWaiter common.StackWaiter) Executor {
	return func() error {
		lbStackName := common.CreateStackName(namespace, common.StackTypeLoadBalancer, environmentName)
		workflow.lbStack = stackWaiter.AwaitFinalStatus(lbStackName)

		envStackName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
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

func (workflow *serviceWorkflow) serviceRolesetUpserter(rolesetUpserter common.RolesetUpserter, rolesetGetter common.RolesetGetter, environmentName string) Executor {
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
		workflow.ecsEventsRoleArn = serviceRoleset["EcsEventsRoleArn"]

		return nil
	}
}

func (workflow *serviceWorkflow) serviceApplyEcsParams(service *common.Service, params map[string]string, rolesetGetter common.RolesetGetter) Executor {
	return func() error {

		params["EcsCluster"] = fmt.Sprintf("%s-EcsCluster", workflow.envStack.Name)
		params["LaunchType"] = fmt.Sprintf("%s-LaunchType", workflow.envStack.Name)
		params["ServiceSubnetIds"] = fmt.Sprintf("%s-InstanceSubnetIds", workflow.envStack.Name)
		params["ServiceSecurityGroup"] = fmt.Sprintf("%s-InstanceSecurityGroup", workflow.envStack.Name)
		params["ElbSecurityGroup"] = fmt.Sprintf("%s-InstanceSecurityGroup", workflow.lbStack.Name)
		params["ImageUrl"] = workflow.serviceImage

		cpu := common.CPUMemorySupport[0]
		if service.CPU != 0 {
			params["ServiceCpu"] = strconv.Itoa(service.CPU)
			for _, cpu = range common.CPUMemorySupport {
				if service.CPU <= cpu.CPU {
					break
				}
			}
		}

		memory := cpu.Memory[0]
		if service.Memory != 0 {
			params["ServiceMemory"] = strconv.Itoa(service.Memory)
			for _, memory = range cpu.Memory {
				if service.Memory <= memory {
					break
				}
			}
		}

		if workflow.isFargateProvider()() {
			params["TaskCpu"] = strconv.Itoa(cpu.CPU)
			params["TaskMemory"] = strconv.Itoa(memory)
		}

		if len(service.Links) > 0 {
			params["Links"] = strings.Join(service.Links, ",")
		}

		// force 'awsvpc' network mode for ecs-fargate
		if strings.EqualFold(string(workflow.envStack.Tags["provider"]), string(common.EnvProviderEcsFargate)) {
			params["TaskNetworkMode"] = "awsvpc"
		} else if service.NetworkMode != "" {
			params["TaskNetworkMode"] = service.NetworkMode
		}
		serviceRoleset, err := rolesetGetter.GetServiceRoleset(workflow.envStack.Tags["environment"], workflow.serviceName)
		if err != nil {
			return err
		}

		params["EcsServiceRoleArn"] = serviceRoleset["EcsServiceRoleArn"]
		params["EcsTaskRoleArn"] = serviceRoleset["EcsTaskRoleArn"]
		params["ApplicationAutoScalingRoleArn"] = serviceRoleset["ApplicationAutoScalingRoleArn"]
		params["ServiceName"] = workflow.serviceName

		return nil
	}
}

func (workflow *serviceWorkflow) serviceApplyEc2Params(params map[string]string, rolesetGetter common.RolesetGetter) Executor {
	return func() error {

		params["AppName"] = workflow.appName
		params["RevisionBucket"] = workflow.appRevisionBucket
		params["RevisionKey"] = workflow.appRevisionKey
		params["RevisionBundleType"] = "zip"

		for _, key := range [...]string{
			"SshAllow",
			"InstanceType",
			"ImageId",
			"ImageOsType",
			"KeyName",
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

		serviceRoleset, err := rolesetGetter.GetServiceRoleset(workflow.envStack.Tags["environment"], workflow.serviceName)
		if err != nil {
			return err
		}

		params["EC2InstanceProfileArn"] = serviceRoleset["EC2InstanceProfileArn"]
		params["CodeDeployRoleArn"] = serviceRoleset["CodeDeployRoleArn"]
		params["ServiceName"] = workflow.serviceName

		return nil
	}
}

func (workflow *serviceWorkflow) serviceApplyCommonParams(namespace string, service *common.Service, params map[string]string, environmentName string, stackWaiter common.StackWaiter, elbRuleLister common.ElbRuleLister, paramGetter common.ParamGetter) Executor {
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

		if service.TargetCPUUtilization != 0 {
			params["TargetCPUUtilization"] = strconv.Itoa(service.TargetCPUUtilization)
		}

		dbStackName := common.CreateStackName(namespace, common.StackTypeDatabase, workflow.serviceName, environmentName)
		dbStack := stackWaiter.AwaitFinalStatus(dbStackName)
		if dbStack != nil {
			params["DatabaseName"] = dbStack.Outputs["DatabaseName"]
			params["DatabaseEndpointAddress"] = dbStack.Outputs["DatabaseEndpointAddress"]
			params["DatabaseEndpointPort"] = dbStack.Outputs["DatabaseEndpointPort"]
			params["DatabaseMasterUsername"] = dbStack.Outputs["DatabaseMasterUsername"]

			dbPass, err := paramGetter.GetParam(fmt.Sprintf("%s-%s", dbStackName, "DatabaseMasterPassword"))
			if err != nil {
				log.Warningf("Unable to get db password: %s", err)
			}
			params["DatabaseMasterPassword"] = dbPass
		}

		svcStackName := common.CreateStackName(namespace, common.StackTypeService, workflow.serviceName, environmentName)
		svcStack := stackWaiter.AwaitFinalStatus(svcStackName)

		if workflow.priority > 0 {
			params["PathListenerRulePriority"] = strconv.Itoa(workflow.priority)
			params["HostListenerRulePriority"] = strconv.Itoa(workflow.priority + 1)
		} else if svcStack != nil && svcStack.Status != "ROLLBACK_COMPLETE" {
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
		if service.MinSize != 0 {
			params["ServiceMinSize"] = strconv.Itoa(service.MinSize)
		}
		if service.MaxSize != 0 {
			params["ServiceMaxSize"] = strconv.Itoa(service.MaxSize)
		}
		if len(service.PathPatterns) > 0 {
			params["PathPattern"] = strings.Join(service.PathPatterns, ",")
		}
		if len(service.HostPatterns) > 0 {
			params["HostPattern"] = strings.Join(service.HostPatterns, ",")
		}
		if service.DeploymentStrategy != "" {
			switch {
			case service.DeploymentStrategy == "blue_green":
				params["MinimumHealthPercent"] = "100"
				params["MaximumHealthPercent"] = "200"
			case service.DeploymentStrategy == "replace":
				params["MinimumHealthPercent"] = "0"
				params["MaximumHealthPercent"] = "100"
			case service.DeploymentStrategy == "rolling":
				params["MinimumHealthPercent"] = "50"
				params["MaximumHealthPercent"] = "100"
			default:
				params["MinimumHealthPercent"] = "100"
				params["MaximumHealthPercent"] = "200"
			}
		}

		return nil
	}
}

func (workflow *serviceWorkflow) serviceEc2Deployer(namespace string, service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {

		log.Noticef("Deploying service '%s' to '%s'", workflow.serviceName, environmentName)

		svcStackName := common.CreateStackName(namespace, common.StackTypeService, workflow.serviceName, environmentName)

		resolveServiceEnvironment(service, environmentName)

		tags := createTagMap(&ServiceTags{
			Service:     workflow.serviceName,
			Environment: environmentName,
			Type:        common.StackTypeService,
			Provider:    workflow.envStack.Outputs["provider"],
			Revision:    workflow.codeRevision,
			Repo:        workflow.repoName,
		})
		err := stackUpserter.UpsertStack(svcStackName, "service-ec2.yml", service, stackParams, tags, workflow.cloudFormationRoleArn)
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

func (workflow *serviceWorkflow) serviceEcsDeployer(namespace string, service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Deploying service '%s' to '%s' from '%s'", workflow.serviceName, environmentName, workflow.serviceImage)

		svcStackName := common.CreateStackName(namespace, common.StackTypeService, workflow.serviceName, environmentName)

		resolveServiceEnvironment(service, environmentName)

		tags := createTagMap(&ServiceTags{
			Service:     workflow.serviceName,
			Environment: environmentName,
			Type:        common.StackTypeService,
			Provider:    workflow.envStack.Outputs["provider"],
			Revision:    workflow.codeRevision,
			Repo:        workflow.repoName,
		})

		err := stackUpserter.UpsertStack(svcStackName, "service-ecs.yml", service, stackParams, tags, workflow.cloudFormationRoleArn)
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
		workflow.microserviceTaskDefinitionArn = stack.Outputs["MicroserviceTaskDefinitionArn"]

		return nil
	}
}

func (workflow *serviceWorkflow) serviceCreateSchedules(namespace string, service *common.Service, environmentName string, stackWaiter common.StackWaiter, stackUpserter common.StackUpserter) Executor {
	return func() error {
		log.Noticef("Creating schedules for service '%s' to '%s'", workflow.serviceName, environmentName)
		for _, schedule := range service.Schedule {
			params := make(map[string]string)

			params["ServiceName"] = workflow.serviceName
			params["EcsCluster"] = fmt.Sprintf("%s-EcsCluster", workflow.envStack.Name)
			params["MicroserviceTaskDefinitionArn"] = workflow.microserviceTaskDefinitionArn
			params["EcsEventsRoleArn"] = workflow.ecsEventsRoleArn

			// these parameters are specific to each of the defined schedules
			params["ScheduleExpression"] = schedule.Expression
			commandBytes, err := json.Marshal(schedule.Command)
			if err != nil {
				return err
			}
			params["ScheduleCommand"] = string(commandBytes)

			scheduleStackName := common.CreateStackName(namespace, common.StackTypeSchedule, workflow.serviceName+"-"+strings.ToLower(schedule.Name), environmentName)
			resolveServiceEnvironment(service, environmentName)

			tags := createTagMap(&ScheduleTags{
				Service:     workflow.serviceName,
				Environment: environmentName,
				Type:        common.StackTypeSchedule,
			})

			err = stackUpserter.UpsertStack(scheduleStackName, "schedule.yml", service, params, tags, workflow.cloudFormationRoleArn)
			if err != nil {
				return err
			}
			log.Debugf("Waiting for stack '%s' to complete", scheduleStackName)
			stack := stackWaiter.AwaitFinalStatus(scheduleStackName)
			if stack == nil {
				return fmt.Errorf("Unable to create stack %s", scheduleStackName)
			}
			if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
				return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
			}
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
