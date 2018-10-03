package workflows

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
)

// NewEnvironmentViewer create a new workflow for showing an environment
func NewEnvironmentViewer(ctx *common.Context, format string, environmentName string, viewTasks bool, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	var environmentViewer func() error
	if format == JSON {
		environmentViewer = workflow.environmentViewerJSON(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, writer)
	} else if format == SHELL {
		environmentViewer = workflow.environmentViewerSHELL(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, writer)
	} else {
		environmentViewer = workflow.environmentViewerCli(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, ctx.InstanceManager, ctx.TaskManager, ctx.KubernetesResourceManagerProvider, viewTasks, writer)
	}

	return newPipelineExecutor(
		environmentViewer,
	)
}

func (workflow *environmentWorkflow) environmentViewerJSON(namespace string, environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, instanceLister common.ClusterInstanceLister, writer io.Writer) Executor {
	return func() error {
		lbStackName := common.CreateStackName(namespace, common.StackTypeLoadBalancer, environmentName)
		lbStack, _ := stackGetter.GetStack(lbStackName)

		clusterStackName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
		clusterStack, _ := stackGetter.GetStack(clusterStackName)

		output := common.JSONOutput{}
		if lbStack != nil {
			output.Values[FirstValueIndex].Key = BaseURLKey
			output.Values[FirstValueIndex].Value = lbStack.Outputs[BaseURLValueKey]
		} else if clusterStack != nil {
			output.Values[FirstValueIndex].Key = BaseURLKey
			output.Values[FirstValueIndex].Value = clusterStack.Outputs[BaseURLValueKey]
		}

		enc := json.NewEncoder(writer)
		enc.Encode(&output)

		return nil
	}
}

func (workflow *environmentWorkflow) environmentViewerSHELL(namespace string, environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, instanceLister common.ClusterInstanceLister, writer io.Writer) Executor {
	return func() error {
		lbStackName := common.CreateStackName(namespace, common.StackTypeLoadBalancer, environmentName)
		lbStack, _ := stackGetter.GetStack(lbStackName)

		clusterStackName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
		clusterStack, _ := stackGetter.GetStack(clusterStackName)

		output := common.JSONOutput{}
		if lbStack != nil {
			output.Values[FirstValueIndex].Key = BaseURLKey
			output.Values[FirstValueIndex].Value = lbStack.Outputs[BaseURLValueKey]
		} else if clusterStack != nil {
			output.Values[FirstValueIndex].Key = BaseURLKey
			output.Values[FirstValueIndex].Value = clusterStack.Outputs[BaseURLValueKey]
		}

		for _, val := range output.Values {
			fmt.Fprintf(writer, "%s=%s\n", val.Key, val.Value)
		}

		return nil
	}
}

func (workflow *environmentWorkflow) environmentViewerCli(namespace string, environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, clusterInstanceLister common.ClusterInstanceLister, instanceLister common.InstanceLister, taskManager common.TaskManager, kubernetesResourceManagerProvider common.KubernetesResourceManagerProvider, viewTasks bool, writer io.Writer) Executor {
	return func() error {
		lbStackName := common.CreateStackName(namespace, common.StackTypeLoadBalancer, environmentName)
		lbStack, err := stackGetter.GetStack(lbStackName)

		clusterStackName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
		clusterStack, err := stackGetter.GetStack(clusterStackName)

		vpcStackName := common.CreateStackName(namespace, common.StackTypeVpc, environmentName)
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

		if clusterStack != nil && clusterStack.Tags["provider"] == string(common.EnvProviderEcs) {
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
		} else if clusterStack != nil && clusterStack.Tags["provider"] == string(common.EnvProviderEks) {

			workflow.environment = &common.Environment{
				Name: environmentName,
			}
			workflow.connectKubernetes(namespace, kubernetesResourceManagerProvider)()

			/*
				var nodes corev1.NodeList
				err = workflow.kubernetesResourceManager.ListResources(context.TODO(), "", &nodes)
				fmt.Printf("%v", err)
				for _, node := range nodes.Items {
					fmt.Printf("name=%q schedulable=%t\n", *node.Metadata.Name, !*node.Spec.Unschedulable)
				}
			*/
		}

		fmt.Fprint(writer, NewLine)
		fmt.Fprintf(writer, HeadNewlineHeader, Bold(ServicesHeader))
		stacks, err := stackLister.ListStacks(common.StackTypeService, namespace)
		if err != nil {
			return err
		}
		table := buildServiceTable(stacks, environmentName, writer)
		table.Render()

		if viewTasks {
			buildContainerTable(namespace, taskManager, stacks, environmentName, writer)
		}

		fmt.Fprint(writer, NewLine)

		return nil
	}
}

func buildContainerTable(namespace string, taskManager common.TaskManager, stacks []*common.Stack, environmentName string, writer io.Writer) {
	for _, stackValues := range stacks {
		if stackValues.Tags[EnvTagKey] != environmentName {
			continue
		}
		doViewTasks(namespace, taskManager, writer, stacks, stackValues.Tags[SvcTagKey])
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
			stackValues.Tags["revision"],
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
