package workflows

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"io"
	"strconv"
	"strings"
)

type environmentWorkflow struct {
	environment *common.Environment
}

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

func colorizeStatus(stackStatus string) string {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	var color func(a ...interface{}) string
	if strings.HasSuffix(stackStatus, "_FAILED") {
		color = red
	} else if strings.HasSuffix(stackStatus, "_COMPLETE") {
		color = green
	} else {
		color = blue
	}
	return color(stackStatus)
}

// NewEnvironmentLister create a new workflow for listing environments
func NewEnvironmentLister(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentLister(ctx.StackManager, writer),
	)
}

func (workflow *environmentWorkflow) environmentLister(stackLister common.StackLister, writer io.Writer) Executor {
	bold := color.New(color.Bold).SprintFunc()

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeCluster)

		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(writer)
		table.SetHeader([]string{"Environment", "Stack", "Status", "Last Update", "Mu Version"})
		table.SetBorder(false)

		for _, stack := range stacks {

			table.Append([]string{
				bold(stack.Tags["environment"]),
				stack.Name,
				fmt.Sprintf("%s %s", colorizeStatus(stack.Status), stack.StatusReason),
				stack.LastUpdateTime.String(),
				stack.Tags["version"],
			})

		}

		table.Render()

		return nil
	}
}

// NewEnvironmentViewer create a new workflow for showing an environment
func NewEnvironmentViewer(ctx *common.Context, environmentName string, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentViewer(environmentName, ctx.StackManager, ctx.ClusterManager, writer),
	)
}

func (workflow *environmentWorkflow) environmentViewer(environmentName string, stackGetter common.StackGetter, instanceLister common.ClusterInstanceLister, writer io.Writer) Executor {
	bold := color.New(color.Bold).SprintFunc()
	return func() error {
		clusterStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		clusterStack, err := stackGetter.GetStack(clusterStackName)
		if err != nil {
			return err
		}

		vpcStackName := common.CreateStackName(common.StackTypeVpc, environmentName)
		vpcStack, _ := stackGetter.GetStack(vpcStackName)

		fmt.Fprintf(writer, "%s:\t%s\n", bold("Environment"), environmentName)
		fmt.Fprintf(writer, "%s:\t%s (%s)\n", bold("Cluster Stack"), clusterStack.Name, colorizeStatus(clusterStack.Status))
		if vpcStack == nil {
			fmt.Fprintf(writer, "%s:\tunmanaged\n", bold("VPC Stack"))
		} else {
			fmt.Fprintf(writer, "%s:\t%s (%s)\n", bold("VPC Stack"), vpcStack.Name, colorizeStatus(vpcStack.Status))
		}

		fmt.Fprintf(writer, "%s:\t%s\n", bold("Base URL"), clusterStack.Outputs["BaseUrl"])

		fmt.Fprintf(writer, "%s:\n", bold("Container Instances"))
		fmt.Fprint(writer, "\n")

		instances, err := instanceLister.ListInstances(clusterStack.Outputs["EcsCluster"])
		if err != nil {
			return err
		}
		table := buildInstanceTable(writer, instances)
		table.Render()
		fmt.Fprint(writer, "\n")

		fmt.Fprintf(writer, "%s:\n", bold("Services"))
		fmt.Fprint(writer, "\n")
		table = buildServiceTable(writer)
		table.Render()

		fmt.Fprint(writer, "\n")

		return nil
	}
}

func buildServiceTable(writer io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Name", "Status"})
	table.SetBorder(false)
	return table
}

func buildInstanceTable(writer io.Writer, instances []*ecs.ContainerInstance) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"EC2 Instance", "Type", "AMI", "AZ", "Connected", "Status", "# Tasks", "CPU Avail", "Mem Avail"})
	table.SetBorder(false)
	for _, instance := range instances {
		instanceType := "???"
		availZone := "???"
		amiID := "???"
		for _, attr := range instance.Attributes {
			switch aws.StringValue(attr.Name) {
			case "ecs.availability-zone":
				availZone = aws.StringValue(attr.Value)
			case "ecs.instance-type":
				instanceType = aws.StringValue(attr.Value)
			case "ecs.ami-id":
				amiID = aws.StringValue(attr.Value)
			}
		}
		var cpuAvail int64
		var memAvail int64
		for _, resource := range instance.RemainingResources {
			switch aws.StringValue(resource.Name) {
			case "CPU":
				cpuAvail = aws.Int64Value(resource.IntegerValue)
			case "MEMORY":
				memAvail = aws.Int64Value(resource.IntegerValue)
			}
		}
		table.Append([]string{
			aws.StringValue(instance.Ec2InstanceId),
			instanceType,
			amiID,
			availZone,
			fmt.Sprintf("%v", aws.BoolValue(instance.AgentConnected)),
			aws.StringValue(instance.Status),
			fmt.Sprintf("%d", aws.Int64Value(instance.RunningTasksCount)),
			fmt.Sprintf("%d", cpuAvail),
			fmt.Sprintf("%d", memAvail),
		})
	}

	return table
}

// NewEnvironmentTerminator create a new workflow for terminating an environment
func NewEnvironmentTerminator(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentEcsTerminator(environmentName, ctx.StackManager, ctx.StackManager),
		workflow.environmentVpcTerminator(environmentName, ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *environmentWorkflow) environmentEcsTerminator(environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		envStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		err := stackDeleter.DeleteStack(envStackName)
		if err != nil {
			return err
		}

		stackWaiter.AwaitFinalStatus(envStackName)
		return nil
	}
}
func (workflow *environmentWorkflow) environmentVpcTerminator(environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		vpcStackName := common.CreateStackName(common.StackTypeVpc, environmentName)
		err := stackDeleter.DeleteStack(vpcStackName)
		if err != nil {
			log.Debugf("Unable to delete VPC, but ignoring error: %v", err)
		}

		stackWaiter.AwaitFinalStatus(vpcStackName)
		return nil
	}
}
