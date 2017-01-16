package workflows

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
	"time"
)

var log = logging.MustGetLogger("environment")

type environmentWorkflow struct {
	environment *common.Environment
}

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
			err = stackUpserter.UpsertStack(vpcStackName, template, nil, buildEnvironmentTags(environment.Name, common.StackTypeVpc))
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
		envStackName := common.CreateStackName(common.StackTypeCluster, environment.Name)

		log.Infof("Upserting ECS environment '%s' ...", environment.Name)
		template, err := templates.NewTemplate("cluster.yml", environment)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(envStackName, template, vpcImportParams, buildEnvironmentTags(environment.Name, common.StackTypeCluster))
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

// NewEnvironmentLister create a new workflow for listing environments
func NewEnvironmentLister(ctx *common.Context) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentLister(ctx.StackManager),
	)
}

func (workflow *environmentWorkflow) environmentLister(stackLister common.StackLister) Executor {
	bold := color.New(color.Bold).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeCluster)

		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Environment","Stack","Status","Last Update","Mu Version"})
		table.SetBorder(false)

		for _, stack := range stacks {
			var color func(a ...interface{}) string
			if strings.HasSuffix(stack.Status, "_FAILED") {
				color = red
			} else if strings.HasSuffix(stack.Status, "_COMPLETE") {
				color = green
			} else {
				color = blue
			}

			lastUpdate,_ := strconv.ParseInt(stack.Tags["lastupdate"], 10, 64)
			tm := time.Unix(lastUpdate, 0)

			table.Append([]string{
				bold(stack.Tags["environment"]),
				stack.Name,
				fmt.Sprintf("%s %s",color(stack.Status),stack.StatusReason),
				tm.String(),
				stack.Tags["version"],
			})

		}

		table.Render()


		return nil
	}
}
