package workflows

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewEnvironmentViewer create a new workflow for showing an environment
func NewEnvironmentViewer(ctx *common.Context, format string, environmentName string, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	var environmentViewer func() error
	if format == common.JSON {
		environmentViewer = workflow.environmentViewerJSON(environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, writer)
	} else {
		environmentViewer = workflow.environmentViewerCli(environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, ctx.TaskManager, writer)
	}

	return newWorkflow(
		environmentViewer,
	)
}

func (workflow *environmentWorkflow) environmentViewerJSON(environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, instanceLister common.ClusterInstanceLister, writer io.Writer) Executor {
	return func() error {
		clusterStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		clusterStack, err := stackGetter.GetStack(clusterStackName)
		if err != nil {
			return err
		}

		output := common.JSONOutput{}
		output.Values[common.FirstValueIndex].Key = common.BaseURLKey
		output.Values[common.FirstValueIndex].Value = clusterStack.Outputs[common.BaseURLValueKey]

		enc := json.NewEncoder(writer)
		enc.Encode(&output)

		return nil
	}
}

func (workflow *environmentWorkflow) environmentViewerCli(environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, instanceLister common.ClusterInstanceLister, taskManager common.TaskManager, writer io.Writer) Executor {
	return func() error {
		clusterStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		clusterStack, err := stackGetter.GetStack(clusterStackName)
		if err != nil {
			return err
		}

		vpcStackName := common.CreateStackName(common.StackTypeVpc, environmentName)
		vpcStack, _ := stackGetter.GetStack(vpcStackName)

		fmt.Fprintf(writer, common.HeaderValueFormat, common.Bold(common.EnvironmentHeader), environmentName)
		fmt.Fprintf(writer, common.StackFormat, common.Bold(common.ClusterStack), clusterStack.Name, colorizeStackStatus(clusterStack.Status))
		if vpcStack == nil {
			fmt.Fprintf(writer, common.UnmanagedStackFormat, common.Bold(common.VPCStack))
		} else {
			fmt.Fprintf(writer, common.StackFormat, common.Bold(common.VPCStack), vpcStack.Name, colorizeStackStatus(vpcStack.Status))
			fmt.Fprintf(writer, common.HeaderValueFormat, common.Bold(common.BastionHost), vpcStack.Outputs[common.BastionHostKey])
		}

		fmt.Fprintf(writer, common.HeaderValueFormat, common.Bold(common.BaseURLHeader), clusterStack.Outputs[common.BaseURLValueKey])
		fmt.Fprintf(writer, common.HeadNewlineHeader, common.Bold(common.ContainerInstances))
		fmt.Fprint(writer, common.NewLine)

		instances, err := instanceLister.ListInstances(clusterStack.Outputs[common.ECSClusterKey])
		if err != nil {
			return err
		}

		table := buildInstanceTable(writer, instances)
		table.Render()

		fmt.Fprint(writer, common.NewLine)
		fmt.Fprintf(writer, common.HeadNewlineHeader, common.Bold(common.ServicesHeader))
		fmt.Fprint(writer, common.NewLine)
		stacks, err := stackLister.ListStacks(common.StackTypeService)
		if err != nil {
			return err
		}
		table = buildServiceTable(stacks, environmentName, writer)
		table.Render()

		buildContainerTable(taskManager, stacks, environmentName, writer)

		fmt.Fprint(writer, common.NewLine)

		return nil
	}
}

func buildContainerTable(taskManager common.TaskManager, stacks []*common.Stack, environmentName string, writer io.Writer) {
	for _, stackValues := range stacks {
		if stackValues.Tags[common.EnvCmd] != environmentName {
			continue
		}
		viewTasks(taskManager, writer, stacks, stackValues.Tags[common.SvcCmd])
	}
}

func buildServiceTable(stacks []*common.Stack, environmentName string, writer io.Writer) *tablewriter.Table {
	table := common.CreateTableSection(writer, common.ServiceTableHeader)

	for _, stackValues := range stacks {
		if stackValues.Tags[common.EnvCmd] != environmentName {
			continue
		}

		table.Append([]string{
			common.Bold(stackValues.Tags[common.SvcCmd]),
			stackValues.Parameters[common.SvcImageURLKey],
			fmt.Sprintf(common.KeyValueFormat, colorizeStackStatus(stackValues.Status), stackValues.StatusReason),
			stackValues.LastUpdateTime.Local().Format(common.LastUpdateTime),
			stackValues.Tags[common.SvcVersionKey],
		})
	}

	return table
}

func buildInstanceTable(writer io.Writer, instances []*ecs.ContainerInstance) *tablewriter.Table {
	table := common.CreateTableSection(writer, common.EnvironmentAMITableHeader)

	for _, instance := range instances {
		instanceType := common.UnknownValue
		availZone := common.UnknownValue
		amiID := common.UnknownValue
		for _, attr := range instance.Attributes {
			switch aws.StringValue(attr.Name) {
			case common.ECSAvailabilityZoneKey:
				availZone = aws.StringValue(attr.Value)
			case common.ECSInstanceTypeKey:
				instanceType = aws.StringValue(attr.Value)
			case common.ECSAMIKey:
				amiID = aws.StringValue(attr.Value)
			}
		}
		var cpuAvail int64
		var memAvail int64
		for _, resource := range instance.RemainingResources {
			switch aws.StringValue(resource.Name) {
			case common.CPU:
				cpuAvail = aws.Int64Value(resource.IntegerValue)
			case common.MEMORY:
				memAvail = aws.Int64Value(resource.IntegerValue)
			}
		}
		table.Append([]string{
			aws.StringValue(instance.Ec2InstanceId),
			instanceType,
			amiID,
			availZone,
			fmt.Sprintf(common.BoolStringFormat, aws.BoolValue(instance.AgentConnected)),
			aws.StringValue(instance.Status),
			fmt.Sprintf(common.IntStringFormat, aws.Int64Value(instance.RunningTasksCount)),
			fmt.Sprintf(common.IntStringFormat, cpuAvail),
			fmt.Sprintf(common.IntStringFormat, memAvail),
		})
	}

	return table
}
