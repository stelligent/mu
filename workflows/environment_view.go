package workflows

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewEnvironmentViewer create a new workflow for showing an environment
func NewEnvironmentViewer(ctx *common.Context, environmentName string, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentViewer(environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, writer),
	)
}

func (workflow *environmentWorkflow) environmentViewer(environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, instanceLister common.ClusterInstanceLister, writer io.Writer) Executor {
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
		fmt.Fprintf(writer, "%s:\t%s (%s)\n", bold("Cluster Stack"), clusterStack.Name, colorizeStackStatus(clusterStack.Status))
		if vpcStack == nil {
			fmt.Fprintf(writer, "%s:\tunmanaged\n", bold("VPC Stack"))
		} else {
			fmt.Fprintf(writer, "%s:\t%s (%s)\n", bold("VPC Stack"), vpcStack.Name, colorizeStackStatus(vpcStack.Status))
			fmt.Fprintf(writer, "%s:\t%s\n", bold("Bastion Host"), vpcStack.Outputs["BastionHost"])
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
		stacks, err := stackLister.ListStacks(common.StackTypeService)
		if err != nil {
			return err
		}
		table = buildServiceTable(stacks, environmentName, writer)
		table.Render()

		fmt.Fprint(writer, "\n")

		return nil
	}
}

func buildServiceTable(stacks []*common.Stack, environmentName string, writer io.Writer) *tablewriter.Table {
	bold := color.New(color.Bold).SprintFunc()

	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Service", "Image", "Status", "Last Update", "Mu Version"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, stack := range stacks {
		if stack.Tags["environment"] != environmentName {
			continue
		}

		table.Append([]string{
			bold(stack.Tags["service"]),
			stack.Parameters["ImageUrl"],
			fmt.Sprintf("%s %s", colorizeStackStatus(stack.Status), stack.StatusReason),
			stack.LastUpdateTime.Local().Format("2006-01-02 15:04:05"),
			stack.Tags["version"],
		})

	}

	return table
}

func buildInstanceTable(writer io.Writer, instances []*ecs.ContainerInstance) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"EC2 Instance", "Type", "AMI", "AZ", "Connected", "Status", "# Tasks", "CPU Avail", "Mem Avail"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)
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
