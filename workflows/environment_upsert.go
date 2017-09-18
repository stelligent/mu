package workflows

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
)

var ecsImagePattern = "amzn-ami-*-amazon-ecs-optimized"
var ec2ImagePattern = "amzn-ami-hvm-*-x86_64-gp2"

// NewEnvironmentUpserter create a new workflow for upserting an environment
func NewEnvironmentUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)
	ecsStackParams := make(map[string]string)
	elbStackParams := make(map[string]string)
	consulStackParams := make(map[string]string)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = ctx.Config.Repo.Slug

	return newPipelineExecutor(
		workflow.environmentFinder(&ctx.Config, environmentName),
		workflow.environmentVpcUpserter(ecsStackParams, elbStackParams, consulStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
		workflow.environmentElbUpserter(ecsStackParams, elbStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
		newConditionalExecutor(workflow.isConsulEnabled(), workflow.environmentConsulUpserter(consulStackParams, ecsStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager), nil),
		workflow.environmentUpserter(ecsStackParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
	)
}

// Find an environment in config, by name and set the reference
func (workflow *environmentWorkflow) environmentFinder(config *common.Config, environmentName string) Executor {

	return func() error {
		for _, e := range config.Environments {
			if strings.EqualFold(e.Name, environmentName) {
				if e.Provider == "" {
					e.Provider = common.EnvProviderEcs
				}
				workflow.environment = &e
				return nil
			}
		}
		return fmt.Errorf("Unable to find environment named '%s' in configuration", environmentName)
	}
}

func (workflow *environmentWorkflow) environmentVpcUpserter(ecsStackParams map[string]string, elbStackParams map[string]string, consulStackParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		vpcStackParams := make(map[string]string)
		var template io.Reader
		var err error

		var vpcStackName string
		if environment.VpcTarget.VpcID == "" {
			log.Debugf("No VpcTarget, so we will upsert the VPC stack that manages the VPC")
			vpcStackName = common.CreateStackName(common.StackTypeVpc, environment.Name)
			overrides := common.GetStackOverrides(vpcStackName)

			// no target VPC, we need to create/update the VPC stack
			template, err = templates.NewTemplate("vpc.yml", environment, overrides)
			if err != nil {
				return err
			}

			if environment.Cluster.InstanceTenancy != "" {
				vpcStackParams["InstanceTenancy"] = environment.Cluster.InstanceTenancy
			}
			if environment.Cluster.SSHAllow != "" {
				vpcStackParams["SshAllow"] = environment.Cluster.SSHAllow
			} else {
				vpcStackParams["SshAllow"] = "0.0.0.0/0"
			}
			if environment.Cluster.KeyName != "" {
				vpcStackParams["BastionKeyName"] = environment.Cluster.KeyName
				vpcStackParams["BastionImageId"], err = imageFinder.FindLatestImageID(ec2ImagePattern)
				if err != nil {
					return err
				}
			}

			vpcStackParams["ElbInternal"] = strconv.FormatBool(environment.Loadbalancer.Internal)
		} else {
			log.Debugf("VpcTarget exists, so we will upsert the VPC stack that references the VPC attributes")
			vpcStackName = common.CreateStackName(common.StackTypeTarget, environment.Name)
			overrides := common.GetStackOverrides(vpcStackName)

			template, err = templates.NewTemplate("vpc-target.yml", environment, overrides)
			if err != nil {
				return err
			}

			// target VPC referenced from config
			vpcStackParams["VpcId"] = environment.VpcTarget.VpcID
			vpcStackParams["ElbSubnetIds"] = strings.Join(environment.VpcTarget.ElbSubnetIds, ",")
			vpcStackParams["InstanceSubnetIds"] = strings.Join(environment.VpcTarget.InstanceSubnetIds, ",")
		}

		log.Noticef("Upserting VPC environment '%s' ...", environment.Name)

		tags, err := concatTagMaps(environment.Tags, buildEnvironmentTags(environment.Name, environment.Provider, common.StackTypeVpc, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(vpcStackName, template, vpcStackParams, tags)
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

		ecsStackParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
		ecsStackParams["InstanceSubnetIds"] = fmt.Sprintf("%s-InstanceSubnetIds", vpcStackName)

		elbStackParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
		elbStackParams["ElbSubnetIds"] = fmt.Sprintf("%s-ElbSubnetIds", vpcStackName)

		consulStackParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
		consulStackParams["InstanceSubnetIds"] = fmt.Sprintf("%s-InstanceSubnetIds", vpcStackName)
		consulStackParams["ElbSubnetIds"] = fmt.Sprintf("%s-ElbSubnetIds", vpcStackName)

		return nil
	}
}

func (workflow *environmentWorkflow) environmentConsulUpserter(consulStackParams map[string]string, ecsStackParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		consulStackName := common.CreateStackName(common.StackTypeConsul, environment.Name)

		log.Noticef("Upserting Consul environment '%s' ...", environment.Name)
		overrides := common.GetStackOverrides(consulStackName)
		template, err := templates.NewTemplate("consul.yml", environment, overrides)
		if err != nil {
			return err
		}

		stackParams := consulStackParams

		if environment.Cluster.SSHAllow != "" {
			stackParams["SshAllow"] = environment.Cluster.SSHAllow
		} else {
			stackParams["SshAllow"] = "0.0.0.0/0"
		}
		if environment.Cluster.KeyName != "" {
			stackParams["KeyName"] = environment.Cluster.KeyName
		}
		if environment.Cluster.HTTPProxy != "" {
			stackParams["HttpProxy"] = environment.Cluster.HTTPProxy
		}
		if environment.Cluster.InstanceType != "" {
			stackParams["InstanceType"] = environment.Cluster.InstanceType
		}
		if environment.Cluster.ImageID != "" {
			stackParams["ImageId"] = environment.Cluster.ImageID
		} else {
			stackParams["ImageId"], err = imageFinder.FindLatestImageID(ecsImagePattern)
			if err != nil {
				return err
			}

		}

		tags, err := concatTagMaps(environment.Tags, buildEnvironmentTags(environment.Name, environment.Provider, common.StackTypeConsul, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(consulStackName, template, stackParams, tags)
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", consulStackName)
		stack := stackWaiter.AwaitFinalStatus(consulStackName)

		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", consulStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		ecsStackParams["ConsulServerAutoScalingGroup"] = stack.Outputs["ConsulServerAutoScalingGroup"]
		ecsStackParams["ConsulRpcClientSecurityGroup"] = stack.Outputs["ConsulRpcClientSecurityGroup"]

		return nil
	}
}

func (workflow *environmentWorkflow) environmentElbUpserter(ecsStackParams map[string]string, elbStackParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		envStackName := common.CreateStackName(common.StackTypeLoadBalancer, environment.Name)

		log.Noticef("Upserting ELB environment '%s' ...", environment.Name)
		overrides := common.GetStackOverrides(envStackName)
		template, err := templates.NewTemplate("elb.yml", environment, overrides)
		if err != nil {
			return err
		}

		stackParams := elbStackParams

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

		stackParams["ElbInternal"] = strconv.FormatBool(environment.Loadbalancer.Internal)
		tags, err := concatTagMaps(environment.Tags, buildEnvironmentTags(environment.Name, environment.Provider, common.StackTypeLoadBalancer, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(envStackName, template, stackParams, tags)
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

		ecsStackParams["ElbSecurityGroup"] = stack.Outputs["ElbInstanceSecurityGroup"]

		return nil
	}
}

func (workflow *environmentWorkflow) environmentUpserter(ecsStackParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Debugf("Using provider '%s' for environment", workflow.environment.Provider)

		environment := workflow.environment
		envStackName := common.CreateStackName(common.StackTypeEnv, environment.Name)

		var templateName string
		var imagePattern string
		if environment.Provider == common.EnvProviderEcs {
			templateName = "env-ecs.yml"
			imagePattern = ecsImagePattern
		} else if environment.Provider == common.EnvProviderEc2 {
			templateName = "env-ec2.yml"
			imagePattern = ec2ImagePattern
		}

		log.Noticef("Upserting environment '%s' ...", environment.Name)
		overrides := common.GetStackOverrides(envStackName)
		template, err := templates.NewTemplate(templateName, environment, overrides)
		if err != nil {
			return err
		}

		stackParams := ecsStackParams

		if environment.Cluster.SSHAllow != "" {
			stackParams["SshAllow"] = environment.Cluster.SSHAllow
		} else {
			stackParams["SshAllow"] = "0.0.0.0/0"
		}
		if environment.Cluster.InstanceType != "" {
			stackParams["InstanceType"] = environment.Cluster.InstanceType
		}
		if environment.Cluster.ImageID != "" {
			stackParams["ImageId"] = environment.Cluster.ImageID
		} else {
			stackParams["ImageId"], err = imageFinder.FindLatestImageID(imagePattern)
			if err != nil {
				return err
			}

		}
		if environment.Cluster.DesiredCapacity != 0 {
			stackParams["DesiredCapacity"] = strconv.Itoa(environment.Cluster.DesiredCapacity)
		}
		if environment.Cluster.MaxSize != 0 {
			stackParams["MaxSize"] = strconv.Itoa(environment.Cluster.MaxSize)
		}
		if environment.Cluster.KeyName != "" {
			stackParams["KeyName"] = environment.Cluster.KeyName
		}
		if environment.Cluster.ScaleInThreshold != 0 {
			stackParams["ScaleInThreshold"] = strconv.Itoa(environment.Cluster.ScaleInThreshold)
		}
		if environment.Cluster.ScaleOutThreshold != 0 {
			stackParams["ScaleOutThreshold"] = strconv.Itoa(environment.Cluster.ScaleOutThreshold)
		}
		if environment.Cluster.HTTPProxy != "" {
			stackParams["HttpProxy"] = environment.Cluster.HTTPProxy
		}

		tags, err := concatTagMaps(environment.Tags, buildEnvironmentTags(environment.Name, environment.Provider, common.StackTypeEnv, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(envStackName, template, stackParams, tags)
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

func buildEnvironmentTags(environmentName string, envProvider common.EnvProvider, stackType common.StackType, codeRevision string, repoName string) map[string]string {
	return map[string]string{
		EnvironmentTags["Type"]:        string(stackType),
		EnvironmentTags["Environment"]: environmentName,
		EnvironmentTags["Provider"]:    string(envProvider),
		EnvironmentTags["Revision"]:    codeRevision,
		EnvironmentTags["Repo"]:        repoName,
	}
}

func concatTagMaps(ymlMap map[string]interface{}, constMap map[string]string) (map[string]string, error) {

	for key := range EnvironmentTags {
		if _, exists := ymlMap[key]; exists {
			return nil, errors.New("Unable to override tag " + key)
		}
	}

	joinedMap := map[string]string{}
	for key, value := range ymlMap {
		joinedMap["demotag"] = "Inserted By Hard Code"
		if str, ok := value.(string); ok {
			log.Noticef(str)
			joinedMap[key] = str
		}
	}
	for key, value := range constMap {
		joinedMap[key] = value
	}

	return joinedMap, nil
}
