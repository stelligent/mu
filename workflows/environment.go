package workflows

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
)

var log = logging.MustGetLogger("environment")

// NewEnvironmentUpserter create a new workflow for upserting an environment
func NewEnvironmentUpserter(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)
	vpcImportParams := make(map[string]string)

	return newWorkflow(
		workflow.environmentFinder(&ctx.Config, environmentName),
		workflow.environmentVpcUpserter(vpcImportParams, ctx.StackManager, ctx.StackManager),
		workflow.environmentEcsUpserter(vpcImportParams, ctx.StackManager, ctx.StackManager),
	)
}

type environmentWorkflow struct {
	environment *common.Environment
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
			vpcStackName := fmt.Sprintf("mu-vpc-%s", environment.Name)

			// no target VPC, we need to create/update the VPC stack
			log.Infof("Upserting VPC environment '%s' ...", environment.Name)
			template, err := templates.NewTemplate("vpc.yml", environment)
			if err != nil {
				return err
			}
			err = stackUpserter.UpsertStack(vpcStackName, template, nil)
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

func (workflow *environmentWorkflow) environmentEcsUpserter(vpcImportParams map[string]string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		environment := workflow.environment
		envStackName := fmt.Sprintf("mu-env-%s", environment.Name)

		log.Infof("Upserting ECS environment '%s' ...", environment.Name)
		template, err := templates.NewTemplate("cluster.yml", environment)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(envStackName, template, vpcImportParams)
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", envStackName)
		stackWaiter.AwaitFinalStatus(envStackName)

		return nil
	}
}
