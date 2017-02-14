package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"io"
	"strconv"
	"strings"
)

var ecsImagePattern = "amzn-ami-*-amazon-ecs-optimized"
var bastionImagePattern = "amzn-ami-hvm-*-x86_64-gp2"

// NewEnvironmentUpserter create a new workflow for upserting an environment
func NewEnvironmentUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)
	vpcImportParams := make(map[string]string)

	return newWorkflow(
		workflow.environmentFinder(&ctx.Config, environmentName),
		workflow.environmentVpcUpserter(vpcImportParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
		workflow.environmentEcsUpserter(vpcImportParams, ctx.StackManager, ctx.StackManager, ctx.StackManager),
	)
}

// Find an environment in config, by name and set the reference
func (workflow *environmentWorkflow) environmentFinder(config *common.Config, environmentName string) Executor {

	return func() error {
		for _, e := range config.Environments {
			if strings.EqualFold(e.Name, environmentName) {
				workflow.environment = &e
				return nil
			}
		}
		return fmt.Errorf("Unable to find environment named '%s' in configuration", environmentName)
	}
}

func (workflow *environmentWorkflow) environmentVpcUpserter(vpcImportParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
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
			}
			if environment.Cluster.KeyName != "" {
				vpcStackParams["BastionKeyName"] = environment.Cluster.KeyName
				vpcStackParams["BastionImageId"], err = imageFinder.FindLatestImageID(bastionImagePattern)
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
			vpcStackParams["EcsSubnetIds"] = strings.Join(environment.VpcTarget.EcsSubnetIds, ",")
		}

		log.Noticef("Upserting VPC environment '%s' ...", environment.Name)
		err = stackUpserter.UpsertStack(vpcStackName, template, vpcStackParams, buildEnvironmentTags(environment.Name, common.StackTypeVpc))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", vpcStackName)
		stackWaiter.AwaitFinalStatus(vpcStackName)

		vpcImportParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
		vpcImportParams["ElbSubnetIds"] = fmt.Sprintf("%s-ElbSubnetIds", vpcStackName)
		vpcImportParams["EcsSubnetIds"] = fmt.Sprintf("%s-EcsSubnetIds", vpcStackName)

		return nil
	}
}

func (workflow *environmentWorkflow) environmentEcsUpserter(vpcImportParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		envStackName := common.CreateStackName(common.StackTypeCluster, environment.Name)

		log.Noticef("Upserting ECS environment '%s' ...", environment.Name)
		overrides := common.GetStackOverrides(envStackName)
		template, err := templates.NewTemplate("cluster.yml", environment, overrides)
		if err != nil {
			return err
		}

		stackParams := vpcImportParams

		if environment.Cluster.SSHAllow != "" {
			stackParams["SshAllow"] = environment.Cluster.SSHAllow
		}
		if environment.Cluster.ImageID != "" {
			stackParams["ImageId"] = environment.Cluster.ImageID
		} else {
			stackParams["ImageId"], err = imageFinder.FindLatestImageID(ecsImagePattern)
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

		if environment.Loadbalancer.Certificate != "" {
			stackParams["ElbCert"] = environment.Loadbalancer.Certificate
		}

		if environment.Loadbalancer.HostedZone != "" {
			stackParams["ElbDomainName"] = environment.Loadbalancer.HostedZone

			if environment.Loadbalancer.Name != "" {
				stackParams["ElbHostName"] = environment.Loadbalancer.Name
			} else {
				stackParams["ElbHostName"] = environment.Name
			}
		}

		stackParams["ElbInternal"] = strconv.FormatBool(environment.Loadbalancer.Internal)

		err = stackUpserter.UpsertStack(envStackName, template, stackParams, buildEnvironmentTags(environment.Name, common.StackTypeCluster))
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", envStackName)
		stackWaiter.AwaitFinalStatus(envStackName)

		return nil
	}
}

func buildEnvironmentTags(environmentName string, stackType common.StackType) map[string]string {
	return map[string]string{
		"type":        string(stackType),
		"environment": environmentName,
	}
}
