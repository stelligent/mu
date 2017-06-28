package workflows

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewEnvironmentViewer create a new workflow for showing an environment
func NewEnvironmentViewer(ctx *common.Context, format string, environmentName string, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	var environmentViewer func() error
	if format == JSON {
		environmentViewer = workflow.environmentViewerJSON(environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, writer)
	} else {
		environmentViewer = workflow.environmentViewerCli(environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, ctx.InstanceManager, ctx.TaskManager, writer)
	}

	return newPipelineExecutor(
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
		output.Values[FirstValueIndex].Key = BaseURLKey
		output.Values[FirstValueIndex].Value = clusterStack.Outputs[BaseURLValueKey]

		enc := json.NewEncoder(writer)
		enc.Encode(&output)

		return nil
	}
}

func (workflow *environmentWorkflow) environmentViewerCli(environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, clusterInstanceLister common.ClusterInstanceLister, instanceLister common.InstanceLister, taskManager common.TaskManager, writer io.Writer) Executor {
	return func() error {
		lbStackName := common.CreateStackName(common.StackTypeLoadBalancer, environmentName)
		lbStack, err := stackGetter.GetStack(lbStackName)

		clusterStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		clusterStack, err := stackGetter.GetStack(clusterStackName)

		vpcStackName := common.CreateStackName(common.StackTypeVpc, environmentName)
		vpcStack, _ := stackGetter.GetStack(vpcStackName)

		fmt.Fprintf(writer, HeaderValueFormat, Bold(EnvironmentHeader), environmentName)
		if clusterStack != nil {
			fmt.Fprintf(writer, StackFormat, Bold(ClusterStack), clusterStack.Name, colorizeStackStatus(clusterStack.Status))
		}

		if vpcStack == nil {
			fmt.Fprintf(writer, UnmanagedStackFormat, Bold(VPCStack))
		} else {
			fmt.Fprintf(writer, StackFormat, Bold(VPCStack), vpcStack.Name, colorizeStackStatus(vpcStack.Status))
			fmt.Fprintf(writer, HeaderValueFormat, Bold(BastionHost), vpcStack.Outputs[BastionHostKey])
		}

		if lbStack != nil {
			fmt.Fprintf(writer, HeaderValueFormat, Bold(BaseURLHeader), lbStack.Outputs[BaseURLValueKey])
		} else if clusterStack != nil {
			fmt.Fprintf(writer, HeaderValueFormat, Bold(BaseURLHeader), clusterStack.Outputs[BaseURLValueKey])
		}

		if clusterStack != nil {
			fmt.Fprintf(writer, HeadNewlineHeader, Bold(ContainerInstances))
			containerInstances, err := clusterInstanceLister.ListInstances(clusterStack.Outputs[ECSClusterKey])
			if err != nil {
				return err
			}

			instanceIds := make([]string, len(containerInstances))
			for i, containerInstance := range containerInstances {
				instanceIds[i] = common.StringValue(containerInstance.Ec2InstanceId)
			}
			instances, err := instanceLister.ListInstances(instanceIds...)
			if err != nil {
				return err
			}

			table := buildInstanceTable(writer, containerInstances, instances)
			table.Render()
		}

		fmt.Fprint(writer, NewLine)
		fmt.Fprintf(writer, HeadNewlineHeader, Bold(ServicesHeader))
		stacks, err := stackLister.ListStacks(common.StackTypeService)
		if err != nil {
			return err
		}
		table := buildServiceTable(stacks, environmentName, writer)
		table.Render()

		buildContainerTable(taskManager, stacks, environmentName, writer)

		fmt.Fprint(writer, NewLine)

		return nil
	}
}

func buildContainerTable(taskManager common.TaskManager, stacks []*common.Stack, environmentName string, writer io.Writer) {
	for _, stackValues := range stacks {
		if stackValues.Tags[EnvTagKey] != environmentName {
			continue
		}
		viewTasks(taskManager, writer, stacks, stackValues.Tags[SvcTagKey])
	}
}

func buildServiceTable(stacks []*common.Stack, environmentName string, writer io.Writer) *tablewriter.Table {
	table := CreateTableSection(writer, ServiceTableHeader)

	for _, stackValues := range stacks {
		if stackValues.Tags[EnvTagKey] != environmentName {
			continue
		}

		table.Append([]string{
			Bold(stackValues.Tags[SvcTagKey]),
			simplifyRepoURL(stackValues.Parameters[SvcImageURLKey]),
			fmt.Sprintf(KeyValueFormat, colorizeStackStatus(stackValues.Status), stackValues.StatusReason),
			stackValues.LastUpdateTime.Local().Format(LastUpdateTime),
		})
	}

	return table
}

func buildInstanceTable(writer io.Writer, containerInstances []common.ContainerInstance, instances []common.Instance) *tablewriter.Table {
	table := CreateTableSection(writer, EnvironmentAMITableHeader)

	instanceIps := make(map[string]string)
	for _, instance := range instances {
		instanceIps[common.StringValue(instance.InstanceId)] = common.StringValue(instance.PrivateIpAddress)
	}

	for _, instance := range containerInstances {
		instanceType := UnknownValue
		availZone := UnknownValue
		amiID := UnknownValue
		for _, attr := range instance.Attributes {
			switch common.StringValue(attr.Name) {
			case ECSAvailabilityZoneKey:
				availZone = common.StringValue(attr.Value)
			case ECSInstanceTypeKey:
				instanceType = common.StringValue(attr.Value)
			case ECSAMIKey:
				amiID = common.StringValue(attr.Value)
			}
		}
		var cpuAvail int64
		var memAvail int64
		for _, resource := range instance.RemainingResources {
			switch common.StringValue(resource.Name) {
			case CPU:
				cpuAvail = common.Int64Value(resource.IntegerValue)
			case MEMORY:
				memAvail = common.Int64Value(resource.IntegerValue)
			}
		}
		table.Append([]string{
			common.StringValue(instance.Ec2InstanceId),
			instanceType,
			amiID,
			instanceIps[common.StringValue(instance.Ec2InstanceId)],
			availZone,
			fmt.Sprintf(BoolStringFormat, common.BoolValue(instance.AgentConnected)),
			common.StringValue(instance.Status),
			fmt.Sprintf(IntStringFormat, common.Int64Value(instance.RunningTasksCount)),
			fmt.Sprintf(IntStringFormat, cpuAvail),
			fmt.Sprintf(IntStringFormat, memAvail),
		})
	}

	return table
}
