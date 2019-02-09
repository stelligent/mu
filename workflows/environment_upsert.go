package workflows

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
)

var ecsImageOwner = "amazon"
var ecsImagePattern = "amzn-ami-*-amazon-ecs-optimized"
var eksImageOwner = "602401143452"
var eksImagePattern = "amazon-eks-node-v*"
var ec2ImageOwner = "amazon"
var ec2ImagePattern = "amzn-ami-hvm-*-x86_64-gp2"

// NewEnvironmentsUpserter create a new workflow for upserting n environments
func NewEnvironmentsUpserter(ctx *common.Context, environmentNames []string) Executor {
	envWorkflows := make([]Executor, len(environmentNames))
	for i, environmentName := range environmentNames {
		envWorkflows[i] = newEnvironmentUpserter(ctx, environmentName)
	}
	return newParallelExecutor(envWorkflows...)
}

func newEnvironmentUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)
	envStackParams := make(map[string]string)
	elbStackParams := make(map[string]string)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = ctx.Config.Repo.Slug

	serviceName := ctx.Config.Service.Name
	if serviceName == "" {
		serviceName = ctx.Config.Repo.Name
	}

	return newPipelineExecutor(
		workflow.environmentFinder(&ctx.Config, environmentName),
		workflow.environmentNormalizer(),
		workflow.environmentRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, envStackParams),
		workflow.environmentVpcUpserter(ctx.Config.Namespace, envStackParams, elbStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.StackManager),
		newConditionalExecutor(workflow.isKubernetesProvider(),
			newPipelineExecutor(
				workflow.environmentKubernetesBootstrapper(ctx.Config.Namespace, envStackParams, ctx.StackManager, ctx.StackManager),
				workflow.environmentUpserter(ctx.Config.Namespace, envStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
				workflow.connectKubernetes(ctx.Config.Namespace, ctx.KubernetesResourceManagerProvider),
				workflow.environmentKubernetesClusterUpserter(ctx.Config.Namespace, serviceName, ctx.Region, ctx.AccountID, ctx.Partition),
				workflow.environmentKubernetesIngressUpserter(ctx.Config.Namespace, ctx.Region, ctx.AccountID, ctx.Partition),
			),
			newPipelineExecutor(
				workflow.environmentElbUpserter(ctx.Config.Namespace, envStackParams, elbStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
				workflow.environmentUpserter(ctx.Config.Namespace, envStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
			),
		),
	)
}

// Find an environment in config, by name and set the reference
func (workflow *environmentWorkflow) environmentFinder(config *common.Config, environmentName string) Executor {

	return func() error {
		for _, e := range config.Environments {
			if strings.EqualFold(e.Name, environmentName) {
				workflow.environment = &e

				workflow.rbacServices = make([]*subjectRoleBinding, 0)
				workflow.rbacUsers = make([]*subjectRoleBinding, 0)
				for _, binding := range config.RBAC {
					if len(binding.Environments) > 0 {
						found := false
						for _, env := range binding.Environments {
							if env == environmentName {
								found = true
								break
							}
						}

						if !found {
							log.Debugf("Skipping binding %v - unable to match env %v", binding, environmentName)
							continue
						}
					}

					for _, service := range binding.Services {
						log.Debugf("Binding service %s to role %s", service, binding.Role)
						workflow.rbacServices = append(workflow.rbacServices, &subjectRoleBinding{
							Name: service,
							Role: string(binding.Role),
						})
					}
					for _, user := range binding.Users {
						log.Debugf("Binding user %s to role %s", user, binding.Role)
						workflow.rbacUsers = append(workflow.rbacUsers, &subjectRoleBinding{
							Name: user,
							Role: string(binding.Role),
						})
					}
				}
				return nil
			}
		}
		return common.Warningf("Unable to find environment named '%s' in configuration", environmentName)
	}
}

func (workflow *environmentWorkflow) environmentVpcUpserter(namespace string,
	envStackParams map[string]string, elbStackParams map[string]string, imageFinder common.ImageFinder,
	stackUpserter common.StackUpserter, stackWaiter common.StackWaiter, azCounter common.AZCounter) Executor {
	return func() error {
		environment := workflow.environment
		vpcStackParams := make(map[string]string)
		var err error

		var vpcStackName string
		var vpcTemplateName string
		if environment.VpcTarget.Environment != "" {
			targetNamespace := common.NewStringIfNotEmpty(namespace, environment.VpcTarget.Namespace)

			log.Debugf("VpcTarget exists for different environment; targeting that VPC")
			vpcStackName = common.CreateStackName(targetNamespace, common.StackTypeVpc, environment.VpcTarget.Environment)
		} else if environment.VpcTarget.VpcID != "" {
			log.Debugf("VpcTarget exists, so we will upsert the VPC stack that references the VPC attributes")
			vpcStackName = common.CreateStackName(namespace, common.StackTypeTarget, environment.Name)
			vpcTemplateName = common.TemplateVPCTarget

			// target VPC referenced from config
			vpcStackParams["VpcId"] = environment.VpcTarget.VpcID
			vpcStackParams["ElbSubnetIds"] = strings.Join(environment.VpcTarget.ElbSubnetIds, ",")
			vpcStackParams["InstanceSubnetIds"] = strings.Join(environment.VpcTarget.InstanceSubnetIds, ",")
		} else {
			log.Debugf("No VpcTarget, so we will upsert the VPC stack that manages the VPC")
			vpcStackName = common.CreateStackName(namespace, common.StackTypeVpc, environment.Name)
			vpcTemplateName = common.TemplateVPC

			common.NewMapElementIfNotEmpty(vpcStackParams, "InstanceTenancy", string(environment.Cluster.InstanceTenancy))

			vpcStackParams["SshAllow"] = "0.0.0.0/0"
			common.NewMapElementIfNotEmpty(vpcStackParams, "SshAllow", environment.Cluster.SSHAllow)

			if environment.Cluster.KeyName != "" {
				vpcStackParams["BastionKeyName"] = environment.Cluster.KeyName
				vpcStackParams["BastionImageId"], err = imageFinder.FindLatestImageID(ec2ImageOwner, ec2ImagePattern)
				if err != nil {
					return err
				}
			}

			vpcStackParams["ElbInternal"] = strconv.FormatBool(environment.Loadbalancer.Internal)

			if environment.Provider == common.EnvProviderEks || environment.Provider == common.EnvProviderEksFargate {
				vpcStackParams["EKSClusterName"] = common.CreateStackName(namespace, common.StackTypeEnv, workflow.environment.Name)
			}
		}

		azCount, err := azCounter.CountAZs()
		if err != nil {
			return err
		}
		if azCount < 2 {
			return fmt.Errorf("Only found %v availability zones...need at least 2", azCount)
		}
		vpcStackParams["AZCount"] = strconv.Itoa(azCount)

		vpcStackParams["Namespace"] = namespace
		vpcStackParams["EnvironmentName"] = environment.Name

		if vpcTemplateName != "" {
			log.Noticef("Upserting VPC environment '%s' ...", environment.Name)

			tags := createTagMap(&EnvironmentTags{
				Environment: environment.Name,
				Type:        string(common.StackTypeVpc),
				Provider:    string(environment.Provider),
				Revision:    workflow.codeRevision,
				Repo:        workflow.repoName,
			})

			err = stackUpserter.UpsertStack(vpcStackName, vpcTemplateName, environment, vpcStackParams, tags, "", workflow.cloudFormationRoleArn)
			if err != nil {
				return err
			}

			log.Debugf("Waiting for stack '%s' to complete", vpcStackName)
			stack := stackWaiter.AwaitFinalStatus(vpcStackName)

			if stack == nil {
				return fmt.Errorf("Unable to create stack %s", vpcStackName)
			}
			if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
				return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
			}
		}

		envStackParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
		envStackParams["InstanceSubnetIds"] = fmt.Sprintf("%s-InstanceSubnetIds", vpcStackName)

		elbStackParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
		elbStackParams["ElbSubnetIds"] = fmt.Sprintf("%s-ElbSubnetIds", vpcStackName)

		return nil
	}
}

func (workflow *environmentWorkflow) environmentRolesetUpserter(rolesetUpserter common.RolesetUpserter, rolesetGetter common.RolesetGetter, envStackParams map[string]string) Executor {
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

		err = rolesetUpserter.UpsertEnvironmentRoleset(workflow.environment.Name)
		if err != nil {
			return err
		}

		environmentRoleset, err := rolesetGetter.GetEnvironmentRoleset(workflow.environment.Name)
		if err != nil {
			return err
		}

		envStackParams["EC2InstanceProfileArn"] = environmentRoleset["EC2InstanceProfileArn"]
		if workflow.environment.Provider == common.EnvProviderEks {
			envStackParams["EksServiceRoleArn"] = environmentRoleset["EksServiceRoleArn"]
		}
		workflow.ec2RoleArn = environmentRoleset["EC2RoleArn"]

		return nil
	}
}

func (workflow *environmentWorkflow) environmentElbUpserter(namespace string, envStackParams map[string]string, elbStackParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		envStackName := common.CreateStackName(namespace, common.StackTypeLoadBalancer, environment.Name)

		log.Noticef("Upserting ELB environment '%s' ...", environment.Name)

		stackParams := elbStackParams
		stackParams["Namespace"] = namespace
		stackParams["EnvironmentName"] = environment.Name

		if environment.Loadbalancer.Certificate != "" {
			stackParams["ElbCert"] = environment.Loadbalancer.Certificate
		}

		if environment.Loadbalancer.HostedZone != "" {
			stackParams["ElbDomainName"] = environment.Loadbalancer.HostedZone

			if environment.Loadbalancer.Name == "" {
				// default to env name
				stackParams["ElbHostName"] = environment.Name
			} else {
				stackParams["ElbHostName"] = environment.Loadbalancer.Name
			}
		}

		if environment.Discovery.Name == "" {
			stackParams["ServiceDiscoveryName"] = fmt.Sprintf("%s.%s.local", environment.Name, namespace)
		} else {
			stackParams["ServiceDiscoveryName"] = environment.Discovery.Name
		}

		stackParams["ElbInternal"] = strconv.FormatBool(environment.Loadbalancer.Internal)

		if environment.Loadbalancer.AccessLogs.S3BucketName != "" {
			stackParams["ElbAccessLogsS3BucketName"] = environment.Loadbalancer.AccessLogs.S3BucketName
		}

		if environment.Loadbalancer.AccessLogs.S3Prefix != "" {
			stackParams["ElbAccessLogsS3Prefix"] = environment.Loadbalancer.AccessLogs.S3Prefix
		}

		tags := createTagMap(&EnvironmentTags{
			Environment: environment.Name,
			Type:        string(common.StackTypeLoadBalancer),
			Provider:    string(environment.Provider),
			Revision:    workflow.codeRevision,
			Repo:        workflow.repoName,
		})

		err := stackUpserter.UpsertStack(envStackName, common.TemplateELB, environment, stackParams, tags, "", workflow.cloudFormationRoleArn)
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", envStackName)
		stack := stackWaiter.AwaitFinalStatus(envStackName)

		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", envStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		envStackParams["ElbSecurityGroup"] = stack.Outputs["ElbInstanceSecurityGroup"]

		return nil
	}
}

func (workflow *environmentWorkflow) environmentUpserter(namespace string, envStackParams map[string]string,
	imageFinder common.ImageFinder, stackUpserter common.StackUpserter,
	stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Debugf("Using provider '%s' for environment", workflow.environment.Provider)

		environment := workflow.environment
		envStackName := common.CreateStackName(namespace, common.StackTypeEnv, environment.Name)

		stackParams := envStackParams
		stackParams["Namespace"] = namespace
		stackParams["EnvironmentName"] = environment.Name

		var templateName string
		var imagePattern string
		var imageOwner string
		envMapping := map[common.EnvProvider]map[string]string{
			common.EnvProviderEcs: map[string]string{
				"templateName": common.TemplateEnvECS,
				"imagePattern": ecsImagePattern,
				"imageOwner":   ecsImageOwner,
				"launchType":   "EC2"},
			common.EnvProviderEcsFargate: map[string]string{
				"templateName": common.TemplateEnvECS,
				"imagePattern": ecsImagePattern,
				"imageOwner":   ecsImageOwner,
				"launchType":   "FARGATE"},
			common.EnvProviderEks: map[string]string{
				"templateName": common.TemplateEnvEKS,
				"imagePattern": eksImagePattern,
				"imageOwner":   eksImageOwner,
				"launchType":   "EC2"},
			/*
				common.EnvProviderEksFargate: map[string]string{
				    "templateName": common.TemplateEnvEKS,
					"imagePattern": eksImagePattern,
					"imageOwner": eksImageOwner,
					"launchType":   "FARGATE"},
			*/
			common.EnvProviderEc2: map[string]string{
				"templateName": common.TemplateEnvEC2,
				"imagePattern": ec2ImagePattern,
				"imageOwner":   ec2ImageOwner}}
		templateName = envMapping[environment.Provider]["templateName"]
		imagePattern = envMapping[environment.Provider]["imagePattern"]
		imageOwner = envMapping[environment.Provider]["imageOwner"]
		common.NewMapElementIfNotEmpty(stackParams, "LaunchType", envMapping[environment.Provider]["launchType"])

		log.Noticef("Upserting environment '%s' ...", environment.Name)

		// Default SshAllow if none defined
		stackParams["SshAllow"] = "0.0.0.0/0"
		common.NewMapElementIfNotEmpty(stackParams, "SshAllow", environment.Cluster.SSHAllow)
		common.NewMapElementIfNotEmpty(stackParams, "InstanceType", environment.Cluster.InstanceType)
		common.NewMapElementIfNotEmpty(stackParams, "ExtraUserData", environment.Cluster.ExtraUserData)
		common.NewMapElementIfNotEmpty(stackParams, "ImageId", environment.Cluster.ImageID)

		if environment.Cluster.ImageID == "" {
			var err error
			stackParams["ImageId"], err = imageFinder.FindLatestImageID(imageOwner, imagePattern)
			if err != nil {
				return err
			}
		}
		common.NewMapElementIfNotEmpty(stackParams, "ImageOsType", environment.Cluster.ImageOsType)

		common.NewMapElementIfNotZero(stackParams, "DesiredCapacity", environment.Cluster.DesiredCapacity)
		common.NewMapElementIfNotZero(stackParams, "MinSize", environment.Cluster.MinSize)
		common.NewMapElementIfNotZero(stackParams, "MaxSize", environment.Cluster.MaxSize)

		common.NewMapElementIfNotEmpty(stackParams, "KeyName", environment.Cluster.KeyName)

		common.NewMapElementIfNotZero(stackParams, "TargetCPUReservation", environment.Cluster.TargetCPUReservation)
		common.NewMapElementIfNotZero(stackParams, "TargetMemoryReservation", environment.Cluster.TargetMemoryReservation)

		common.NewMapElementIfNotEmpty(stackParams, "HttpProxy", environment.Cluster.HTTPProxy)

		tags := createTagMap(&EnvironmentTags{
			Environment: environment.Name,
			Type:        string(common.StackTypeEnv),
			Provider:    string(environment.Provider),
			Revision:    workflow.codeRevision,
			Repo:        workflow.repoName,
		})

		err := stackUpserter.UpsertStack(envStackName, templateName, environment, stackParams, tags, "", workflow.cloudFormationRoleArn)
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", envStackName)
		stack := stackWaiter.AwaitFinalStatus(envStackName)

		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", envStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}
}

func (workflow *environmentWorkflow) environmentKubernetesBootstrapper(namespace string, envStackParams map[string]string, stackWaiter common.StackWaiter, stackUpserter common.StackUpserter) Executor {
	return func() error {
		envStackName := common.CreateStackName(namespace, common.StackTypeEnv, workflow.environment.Name)
		envStack := stackWaiter.AwaitFinalStatus(envStackName)

		if envStack == nil || envStack.Status == cloudformation.StackStatusRollbackComplete {
			log.Debugf("Attempting to bootstrap stack '%s'", envStackName)

			stackParams := make(map[string]string)
			stackParams["VpcId"] = envStackParams["VpcId"]
			stackParams["InstanceSubnetIds"] = envStackParams["InstanceSubnetIds"]
			stackParams["EksServiceRoleArn"] = envStackParams["EksServiceRoleArn"]
			stackParams["Namespace"] = namespace
			stackParams["EnvironmentName"] = workflow.environment.Name

			tags := createTagMap(&EnvironmentTags{
				Environment: workflow.environment.Name,
				Type:        string(common.StackTypeEnv),
				Provider:    string(workflow.environment.Provider),
				Revision:    workflow.codeRevision,
				Repo:        workflow.repoName,
			})

			err := stackUpserter.UpsertStack(envStackName, common.TemplateEnvEKSBootstrap, workflow.environment, stackParams, tags, "", "")
			if err != nil {
				return err
			}
			log.Debugf("Waiting for stack '%s' to complete", envStackName)
			stack := stackWaiter.AwaitFinalStatus(envStackName)

			if stack == nil {
				return fmt.Errorf("Unable to create stack %s", envStackName)
			}
			if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
				return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
			}
		} else {
			log.Debugf("Stack '%s' has already been bootstrapped", envStackName)
		}
		return nil
	}
}

func (workflow *environmentWorkflow) environmentKubernetesClusterUpserter(namespace string, serviceName string, region string, accountID string, partition string) Executor {
	return func() error {

		templateData := map[string]interface{}{
			"EC2RoleArn":   workflow.ec2RoleArn,
			"ServiceName":  serviceName,
			"AWSAccountId": accountID,
			"MuNamespace":  namespace,
			"MuVersion":    common.GetVersion(),
			"AWSRegion":    region,
			"AWSPartition": partition,
			"RBACServices": workflow.rbacServices,
			"RBACUsers":    workflow.rbacUsers,
		}

		clusterName := common.CreateStackName(namespace, common.StackTypeEnv, workflow.environment.Name)
		log.Noticef("Upserting kubernetes cluster '%s' ...", clusterName)

		return workflow.kubernetesResourceManager.UpsertResources(common.TemplateK8sCluster, templateData)
	}
}

func (workflow *environmentWorkflow) environmentKubernetesIngressUpserter(namespace string, region string, accountID string, partition string) Executor {
	return func() error {

		var elbCertArn string
		if workflow.environment.Loadbalancer.Certificate != "" {
			partition := partition
			region := region
			accountID := accountID
			elbCertArn = fmt.Sprintf("arn:%s:acm:%s:%s:certificate/%s", partition, region, accountID, workflow.environment.Loadbalancer.Certificate)
		}
		templateData := map[string]interface{}{
			"Namespace":  "mu-ingress",
			"ElbCertArn": elbCertArn,
		}

		clusterName := common.CreateStackName(namespace, common.StackTypeEnv, workflow.environment.Name)
		log.Noticef("Upserting kubernetes ingress in cluster '%s' ...", clusterName)

		return workflow.kubernetesResourceManager.UpsertResources(common.TemplateK8sIngress, templateData)
	}
}
