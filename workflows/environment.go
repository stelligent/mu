package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
)

// NewEnvironmentUpserter create a new workflow for upserting an environment
func NewEnvironmentUpserter(ctx *common.Context, environmentName string) Executor {

	var environment *common.Environment
	var vpcImportParams = make(map[string]string)

	return newWorkflow(
		environmentFinder(&environment, &ctx.Config, environmentName),
		environmentVpcUpserter(environment, vpcImportParams, ctx.StackManager, ctx.StackManager),
		environmentEcsUpserter(environment, vpcImportParams, ctx.StackManager, ctx.StackManager),
	)
}

// Find an environment in config, by name and set the reference
func environmentFinder(environment **common.Environment, config *common.Config, environmentName string) Executor {

	return func() error {
		for _, e := range config.Environments {
			if strings.EqualFold(e.Name, environmentName) {
				*environment = &e
				return nil
			}
		}
		return fmt.Errorf("Unable to find environment named '%s' in configuration", environmentName)
	}
}

func environmentVpcUpserter(environment *common.Environment, vpcImportParams map[string]string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		if environment.VpcTarget.VpcID == "" {
			vpcStackName := fmt.Sprintf("mu-vpc-%s", environment.Name)

			// no target VPC, we need to create/update the VPC stack
			fmt.Printf("upserting VPC environment:%s stack:%s\n", environment.Name, vpcStackName)
			template, err := templates.NewTemplate("cluster.yml", environment)
			if err != nil {
				return err
			}
			err = stackUpserter.UpsertStack(vpcStackName, template, nil)
			if err != nil {
				return err
			}

			stackWaiter.AwaitFinalStatus(vpcStackName)

			// apply default parameters since we manage the VPC
			vpcImportParams["VpcId"] = fmt.Sprintf("%s-VpcId", vpcStackName)
			vpcImportParams["PublicSubnetAZ1Id"] = fmt.Sprintf("%s-PublicSubnetAZ1Id", vpcStackName)
			vpcImportParams["PublicSubnetAZ2Id"] = fmt.Sprintf("%s-PublicSubnetAZ2Id", vpcStackName)
			vpcImportParams["PublicSubnetAZ3Id"] = fmt.Sprintf("%s-PublicSubnetAZ3Id", vpcStackName)
		} else {
			// target VPC referenced from config
			vpcImportParams["VpcId"] = environment.VpcTarget.VpcID
			for index, subnet := range environment.VpcTarget.PublicSubnetIds {
				vpcImportParams[fmt.Sprintf("PublicSubnetAZ%dId", index+1)] = subnet
			}
		}

		return nil
	}
}

func environmentEcsUpserter(environment *common.Environment, vpcImportParams map[string]string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		envStackName := fmt.Sprintf("mu-env-%s", environment.Name)

		fmt.Printf("upserting ECS environment:%s stack:%s\n", environment.Name, envStackName)
		template, err := templates.NewTemplate("cluster.yml", environment)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(envStackName, template, vpcImportParams)
		if err != nil {
			return err
		}
		stackWaiter.AwaitFinalStatus(envStackName)

		return nil
	}
}
