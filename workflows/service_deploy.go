package workflows

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
			), nil),
		newConditionalExecutor(workflow.isEc2Provider(),
			newPipelineExecutor(
				workflow.serviceBucketUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, environmentName),
				workflow.serviceAppUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceApplyEc2Params(stackParams, ctx.RolesetManager),
				workflow.serviceEc2Deployer(ctx.Config.Namespace, &ctx.Config.Service, stackParams, environmentName, ctx.StackManager, ctx.StackManager),
				// TODO - placeholder for doing serviceCreateSchedules for EC2, leaving out-of-scope per @cplee
			), nil),
		newConditionalExecutor(workflow.isEksProvider(),
			newPipelineExecutor(
				workflow.serviceRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, environmentName),
				workflow.serviceRepoUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.connectKubernetes(ctx.KubernetesResourceManagerProvider),
				workflow.serviceEksDBSecret(ctx.Config.Namespace, &ctx.Config.Service, stackParams, environmentName),
				workflow.serviceEksDeployer(ctx.Config.Namespace, &ctx.Config.Service, stackParams, environmentName),
				// TODO - placeholder for doing serviceCreateSchedules for EKS, leaving out-of-scope
			), nil),
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

func checkPriorityNotInUse(elbRuleLister common.ElbRuleLister, listenerArn string, priorities []int) error {
	rules, err := elbRuleLister.ListRules(listenerArn)
	if err != nil {
		return fmt.Errorf("Error checking priorities for listener '%s': %v", listenerArn, err)
	}
	res := []string{}
	for _, rule := range rules {
		priority, _ := strconv.Atoi(common.StringValue(rule.Priority))
		for _, p := range priorities {
			if priority == p {
				res = append(res, strconv.Itoa(p))
			}
		}
	}
	if len(res) > 0 {
		return fmt.Errorf("ELB priority already in use: %s\nChange or remove the priority definition in mu.yml", strings.Join(res, ","))
	}
	return nil
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

		err = rolesetUpserter.UpsertServiceRoleset(environmentName, workflow.serviceName, workflow.appRevisionBucket, workflow.databaseName)
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

func matchRequestedCPU(serviceCPU int, defaultCPU common.CPUMemory) common.CPUMemory {
	for _, cpu := range common.CPUMemorySupport {
		if serviceCPU <= cpu.CPU {
			return cpu
		}
	}
	return defaultCPU
}

func matchRequestedMemory(serviceMemory int, cpu common.CPUMemory, defaultMemory int) int {
	for _, memory := range cpu.Memory {
		if serviceMemory <= memory {
			return memory
		}
	}
	return defaultMemory
}

func getMinMaxPercentForStrategy(deploymentStrategy common.DeploymentStrategy) (string, string) {
	var minHealthyPercent, maxPercent string
	switch deploymentStrategy {
	case common.BlueGreenDeploymentStrategy:
		minHealthyPercent = "100"
		maxPercent = "200"
	case common.ReplaceDeploymentStrategy:
		minHealthyPercent = "0"
		maxPercent = "100"
	case common.RollingDeploymentStrategy:
		minHealthyPercent = "50"
		maxPercent = "100"
	default:
		minHealthyPercent = "100"
		maxPercent = "200"
	}
	return minHealthyPercent, maxPercent
}

// getMaxUnavilableAndSurgeForKubernetesStrategy returns the k8s deployment
// maxUnavailable and maxSurge values for deployment strategies blue/green and
// rolling. For more information,
// see https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-update-deployment.
//
// See common/types.go:type DeploymentStrategy string definition for valid values of
// deploymentStrategy.
//
// maxUnavailable defines the percentage of pods that can be taken out of service.
// maxSurge defines the percentage of extra pods allowed.
func getMaxUnavilableAndSurgePercentForKubernetesStrategy(
	deploymentStrategy common.DeploymentStrategy) (
	maxUnavailable string, maxSurge string) {

	switch deploymentStrategy {
	case common.BlueGreenDeploymentStrategy:
		maxUnavailable = "0"
		maxSurge = "100"
	case common.RollingDeploymentStrategy:
		maxUnavailable = "50"
		maxSurge = "0"
	case common.ReplaceDeploymentStrategy: // not actually used but left for illustration
		maxUnavailable = "100"
		maxSurge = "0"
	default:
		maxUnavailable = "0"
		maxSurge = "100"
	}
	return maxUnavailable, maxSurge
}

func (workflow *serviceWorkflow) serviceApplyEcsParams(service *common.Service, params map[string]string, rolesetGetter common.RolesetGetter) Executor {
	return func() error {

		params["EcsCluster"] = fmt.Sprintf("%s-EcsCluster", workflow.envStack.Name)
		params["LaunchType"] = fmt.Sprintf("%s-LaunchType", workflow.envStack.Name)
		params["ServiceSubnetIds"] = fmt.Sprintf("%s-InstanceSubnetIds", workflow.envStack.Name)
		params["ServiceSecurityGroup"] = fmt.Sprintf("%s-InstanceSecurityGroup", workflow.envStack.Name)
		params["ElbSecurityGroup"] = fmt.Sprintf("%s-InstanceSecurityGroup", workflow.lbStack.Name)
		params["ServiceDiscoveryId"] = fmt.Sprintf("%s-ServiceDiscoveryId", workflow.lbStack.Name)
		params["ServiceDiscoveryName"] = fmt.Sprintf("%s-ServiceDiscoveryName", workflow.lbStack.Name)
		common.NewMapElementIfNotEmpty(params, "ServiceDiscoveryTTL", service.DiscoveryTTL)

		params["ImageUrl"] = workflow.serviceImage

		cpu := common.CPUMemorySupport[0]
		if service.CPU != 0 {
			params["ServiceCpu"] = strconv.Itoa(service.CPU)
			cpu = matchRequestedCPU(service.CPU, cpu)
		}

		memory := cpu.Memory[0]
		if service.Memory != 0 {
			params["ServiceMemory"] = strconv.Itoa(service.Memory)
			memory = matchRequestedMemory(service.Memory, cpu, memory)
		}

		if workflow.isFargateProvider()() {
			params["TaskCpu"] = strconv.Itoa(cpu.CPU)
			params["TaskMemory"] = strconv.Itoa(memory)
		}

		if len(service.Links) > 0 {
			params["Links"] = strings.Join(service.Links, ",")
		}

		params["AssignPublicIp"] = strconv.FormatBool(service.AssignPublicIP)

		// force 'awsvpc' network mode for ecs-fargate
		if strings.EqualFold(string(workflow.envStack.Tags["provider"]), string(common.EnvProviderEcsFargate)) {
			params["TaskNetworkMode"] = common.NetworkModeAwsVpc
		} else if service.NetworkMode != "" {
			params["TaskNetworkMode"] = string(service.NetworkMode)
		}

		serviceRoleset, err := rolesetGetter.GetServiceRoleset(workflow.envStack.Tags["environment"], workflow.serviceName)
		if err != nil {
			return err
		}

		params["EcsServiceRoleArn"] = serviceRoleset["EcsServiceRoleArn"]
		params["EcsTaskRoleArn"] = serviceRoleset["EcsTaskRoleArn"]
		params["ApplicationAutoScalingRoleArn"] = serviceRoleset["ApplicationAutoScalingRoleArn"]
		params["ServiceName"] = workflow.serviceName

		params["MinimumHealthyPercent"], params["MaximumPercent"] = getMinMaxPercentForStrategy(service.DeploymentStrategy)

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
			"ElbSecurityGroup",
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

func (workflow *serviceWorkflow) serviceApplyCommonParams(namespace string, service *common.Service,
	params map[string]string, environmentName string, stackWaiter common.StackWaiter,
	elbRuleLister common.ElbRuleLister, paramGetter common.ParamGetter) Executor {
	return func() error {
		params["VpcId"] = fmt.Sprintf("%s-VpcId", workflow.envStack.Name)

		nextAvailablePriority := 0
		if workflow.lbStack != nil {
			if workflow.lbStack.Outputs["ElbHttpListenerArn"] != "" {
				params["ElbHttpListenerArn"] = fmt.Sprintf("%s-ElbHttpListenerArn", workflow.lbStack.Name)
				if workflow.priority < 1 {
					nextAvailablePriority = 1 + getMaxPriority(elbRuleLister, workflow.lbStack.Outputs["ElbHttpListenerArn"])
				}
			}
			if workflow.lbStack.Outputs["ElbHttpsListenerArn"] != "" {
				params["ElbHttpsListenerArn"] = fmt.Sprintf("%s-ElbHttpsListenerArn", workflow.lbStack.Name)
				if workflow.priority < 1 && nextAvailablePriority == 0 {
					nextAvailablePriority = 1 + getMaxPriority(elbRuleLister, workflow.lbStack.Outputs["ElbHttpsListenerArn"])
				}
			}
		}

		common.NewMapElementIfNotZero(params, "TargetCPUUtilization", service.TargetCPUUtilization)

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
			// make sure manually specified priority is not already in use
			err := checkPriorityNotInUse(elbRuleLister, workflow.lbStack.Outputs["ElbHttpListenerArn"], []int{workflow.priority, workflow.priority + 1})
			if err != nil {
				return err
			}

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

		params["Namespace"] = namespace
		params["EnvironmentName"] = environmentName
		params["ServiceName"] = workflow.serviceName
		common.NewMapElementIfNotZero(params, "ServicePort", service.Port)
		common.NewMapElementIfNotEmpty(params, "ServiceProtocol", string(service.Protocol))
		common.NewMapElementIfNotEmpty(params, "ServiceHealthEndpoint", service.HealthEndpoint)
		common.NewMapElementIfNotZero(params, "ServiceDesiredCount", service.DesiredCount)
		common.NewMapElementIfNotZero(params, "ServiceMinSize", service.MinSize)
		common.NewMapElementIfNotZero(params, "ServiceMaxSize", service.MaxSize)

		if len(service.PathPatterns) > 0 {
			params["PathPattern"] = strings.Join(service.PathPatterns, ",")
		}
		if len(service.HostPatterns) > 0 {
			params["HostPattern"] = strings.Join(service.HostPatterns, ",")
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
		err := stackUpserter.UpsertStack(svcStackName, common.TemplateServiceEC2, service, stackParams, tags, "", workflow.cloudFormationRoleArn)
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

		err := stackUpserter.UpsertStack(svcStackName, common.TemplateServiceECS, service, stackParams, tags, "", workflow.cloudFormationRoleArn)
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

func (workflow *serviceWorkflow) serviceEksDBSecret(namespace string, service *common.Service, stackParams map[string]string, environmentName string) Executor {
	return func() error {
		if stackParams["DatabaseName"] == "" {
			return nil
		}
		log.Noticef("Deploying database secrets for '%s' in '%s'", workflow.serviceName, environmentName)

		params := make(map[string]string)
		params["ServiceName"] = workflow.serviceName
		params["Namespace"] = fmt.Sprintf("mu-service-%s", workflow.serviceName)
		params["Revision"] = workflow.codeRevision
		params["MuVersion"] = common.GetVersion()

		paramKeysToCopy := []string{"DatabaseName", "DatabaseEndpointAddress", "DatabaseEndpointPort", "DatabaseMasterUsername", "DatabaseMasterPassword"}
		for _, key := range paramKeysToCopy {
			params[key] = base64.StdEncoding.EncodeToString([]byte(stackParams[key]))
		}

		return workflow.kubernetesResourceManager.UpsertResources(common.TemplateK8sDatabase, params)
	}
}

// serviceEksDeployer accepts a service and its information and upserts a kubernetes Pod file to
// a k8s cluster
func (workflow *serviceWorkflow) serviceEksDeployer(namespace string, service *common.Service, stackParams map[string]string, environmentName string) Executor {
	return func() error {
		log.Noticef("Deploying service '%s' to '%s' from '%s'", workflow.serviceName, environmentName, workflow.serviceImage)

		servicePort := 8080
		if service.Port != 0 {
			servicePort = service.Port
		}

		serviceProto := common.ServiceProtocolHTTP
		if service.Protocol != "" {
			serviceProto = string(service.Protocol)
		}

		serviceHealthEndpoint := "/health"
		if service.HealthEndpoint != "" {
			serviceHealthEndpoint = service.HealthEndpoint
		}

		pathPatterns := service.PathPatterns
		for idx, pattern := range pathPatterns {
			pathPatterns[idx] = strings.TrimRight(pattern, "*")
		}

		resolveServiceEnvironment(service, environmentName)
		templateData := map[string]interface{}{
			"Namespace":             fmt.Sprintf("mu-service-%s", workflow.serviceName),
			"ServiceName":           workflow.serviceName,
			"ServicePort":           servicePort,
			"ServiceProto":          strings.ToLower(serviceProto),
			"PathPatterns":          pathPatterns,
			"HostPatterns":          service.HostPatterns,
			"ImageUrl":              workflow.serviceImage,
			"ServiceHealthEndpoint": serviceHealthEndpoint,
			"ServiceHealthProto":    strings.ToUpper(serviceProto),
			"Revision":              workflow.codeRevision,
			"MuVersion":             common.GetVersion(),
			"EnvVariables":          service.Environment,
			"DeploymentStrategy":    string(service.DeploymentStrategy),
		}
		// see common/types.go DeploymentStrategy types for valid string values
		templateData["MaxUnavailable"], templateData["MaxSurge"] = getMaxUnavilableAndSurgePercentForKubernetesStrategy(service.DeploymentStrategy)

		if stackParams["DatabaseName"] != "" {
			templateData["DatabaseSecretName"] = fmt.Sprintf("%s-database", workflow.serviceName)
		}

		return workflow.kubernetesResourceManager.UpsertResources(common.TemplateK8sDeployment, templateData)
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

			err = stackUpserter.UpsertStack(scheduleStackName, common.TemplateSchedule, service, params, tags, "", workflow.cloudFormationRoleArn)
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
