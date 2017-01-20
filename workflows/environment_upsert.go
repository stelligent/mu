package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strconv"
	"strings"
)

var ecsImagePattern = "amzn-ami-*-amazon-ecs-optimized"

// NewEnvironmentUpserter create a new workflow for upserting an environment
func NewEnvironmentUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)
	vpcImportParams := make(map[string]string)

	return newWorkflow(
		workflow.environmentFinder(&ctx.Config, environmentName),
		workflow.environmentVpcUpserter(vpcImportParams, ctx.StackManager, ctx.StackManager),
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

func (workflow *environmentWorkflow) environmentVpcUpserter(vpcImportParams map[string]string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		if environment.VpcTarget.VpcID == "" {
			log.Debugf("No VpcTarget, so we will upsert the VPC stack")
			vpcStackName := common.CreateStackName(common.StackTypeVpc, environment.Name)

			// no target VPC, we need to create/update the VPC stack
			log.Infof("Upserting VPC environment '%s' ...", environment.Name)
			template, err := templates.NewTemplate("vpc.yml", environment)
			if err != nil {
				return err
			}

			vpcStackParams := make(map[string]string)
			if environment.Cluster.InstanceTenancy != "" {
				vpcStackParams["InstanceTenancy"] = environment.Cluster.InstanceTenancy
			}
			if environment.Cluster.SSHAllow != "" {
				vpcStackParams["SshAllow"] = environment.Cluster.SSHAllow
			}
			err = stackUpserter.UpsertStack(vpcStackName, template, vpcStackParams, buildEnvironmentTags(environment.Name, common.StackTypeVpc))
			if err != nil {
				return err
			}

			log.Debugf("Waiting for stack '%s' to complete", vpcStackName)
			stackWaiter.AwaitFinalStatus(vpcStackName)

			// apply default parameters since we manage the VPC
			vpcImportParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
			vpcImportParams["PublicSubnetAZ1Id"] = fmt.Sprintf("%s-PublicSubnetAZ1Id", vpcStackName)
			vpcImportParams["PublicSubnetAZ2Id"] = fmt.Sprintf("%s-PublicSubnetAZ2Id", vpcStackName)
			vpcImportParams["PublicSubnetAZ3Id"] = fmt.Sprintf("%s-PublicSubnetAZ3Id", vpcStackName)
		} else {
			log.Debugf("VpcTarget exists, so we will reference the VPC stack")
			// target VPC referenced from config
			vpcImportParams["VpcId"] = environment.VpcTarget.VpcID
			for index, subnet := range environment.VpcTarget.PublicSubnetIds {
				vpcImportParams[fmt.Sprintf("PublicSubnetAZ%dId", index+1)] = subnet
			}
		}

		return nil
	}
}

func (workflow *environmentWorkflow) environmentEcsUpserter(vpcImportParams map[string]string, imageFinder common.ImageFinder, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		envStackName := common.CreateStackName(common.StackTypeCluster, environment.Name)

		log.Infof("Upserting ECS environment '%s' ...", environment.Name)
		template, err := templates.NewTemplate("cluster.yml", environment)
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
			stackParams["Keyname"] = environment.Cluster.KeyName
		}
		if environment.Cluster.ScaleInThreshold != 0 {
			stackParams["ScaleInThreshold"] = strconv.Itoa(environment.Cluster.ScaleInThreshold)
		}
		if environment.Cluster.ScaleOutThreshold != 0 {
			stackParams["ScaleOutThreshold"] = strconv.Itoa(environment.Cluster.ScaleOutThreshold)
		}

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
