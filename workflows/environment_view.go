package workflows

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/stelligent/mu/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// InstanceView representation of instance
type instanceView struct {
	instanceID   string
	instanceType string
	amiID        string
	instanceIP   string
	availZone    string
	ready        bool
	status       string
	taskCount    int64
	cpuAvail     int64
	memAvail     int64
}

// ServiceView representation of service
type serviceView struct {
	name         string
	revision     string
	status       string
	statusReason string
	lastUpdate   string
}

// EnvironmentView representation of environment
type environmentView struct {
	name          string
	provider      common.EnvProvider
	clusterName   string
	clusterStatus string
	vpcName       string
	vpcStatus     string
	bastionHost   string
	baseURL       string
	instances     []*instanceView
	services      []*serviceView
}

// NewEnvironmentViewer create a new workflow for showing an environment
func NewEnvironmentViewer(ctx *common.Context, format string, environmentName string, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)
	view := new(environmentView)

	var environmentViewer func() error
	if format == JSON {
		environmentViewer = workflow.environmentViewerJSON(view, writer)
	} else if format == SHELL {
		environmentViewer = workflow.environmentViewerSHELL(view, writer)
	} else {
		environmentViewer = workflow.environmentViewerCLI(view, writer)
	}

	return newPipelineExecutor(
		workflow.environmentLoader(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager, ctx.ClusterManager, ctx.InstanceManager, ctx.TaskManager, ctx.KubernetesResourceManagerProvider, view),
		environmentViewer,
	)
}

func (workflow *environmentWorkflow) environmentLoader(namespace string, environmentName string, stackGetter common.StackGetter, stackLister common.StackLister, clusterInstanceLister common.ClusterInstanceLister, instanceLister common.InstanceLister, taskManager common.TaskManager, k8sProvider common.KubernetesResourceManagerProvider, view *environmentView) Executor {
	return func() error {
		lbStackName := common.CreateStackName(namespace, common.StackTypeLoadBalancer, environmentName)
		lbStack, _ := stackGetter.GetStack(lbStackName)

		clusterStackName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
		clusterStack, err := stackGetter.GetStack(clusterStackName)
		if err != nil {
			return err
		}

		vpcStackName := common.CreateStackName(namespace, common.StackTypeVpc, environmentName)
		vpcStack, _ := stackGetter.GetStack(vpcStackName)

		workflow.environment = &common.Environment{
			Name:     environmentName,
			Provider: common.EnvProvider(clusterStack.Tags["provider"]),
		}
		kubernetesResourceManager, err := k8sProvider.GetResourceManager(clusterStackName)
		workflow.kubernetesResourceManager = kubernetesResourceManager

		view.name = environmentName
		view.provider = common.EnvProvider(clusterStack.Tags["provider"])
		view.clusterName = clusterStackName
		view.clusterStatus = clusterStack.Status
		view.vpcName = vpcStackName
		view.vpcStatus = vpcStack.Status
		view.bastionHost = vpcStack.Outputs[BastionHostKey]

		if clusterStack.Tags["provider"] == string(common.EnvProviderEks) {
			ingressList, err := workflow.kubernetesResourceManager.ListResources("v1", "Service", "mu-ingress")
			for _, ingress := range ingressList.Items {
				if common.MapGetString(ingress.Object, "metadata", "name") == "nginx-ingress-service" {
					host := common.MapGetString(ingress.Object, "status", "loadBalancer", "ingress", 0, "hostname")
					proto := common.MapGetString(ingress.Object, "spec", "ports", 0, "name")
					view.baseURL = fmt.Sprintf("%s://%s", proto, host)
				}
			}

			nodes, err := workflow.kubernetesResourceManager.ListResources("v1", "Node", "")
			if err != nil {
				return err
			}
			view.instances = buildInstanceViewForEKS(nodes)
		} else {
			if lbStack != nil {
				view.baseURL = lbStack.Outputs[BaseURLValueKey]
			} else {
				view.baseURL = clusterStack.Outputs[BaseURLValueKey]
			}
			if clusterStack.Tags["provider"] == string(common.EnvProviderEcs) {
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

				view.instances = buildInstanceViewForECS(containerInstances, instances)
			}

			view.services, err = buildServiceViewForECS(namespace, "", environmentName, stackLister)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func (workflow *environmentWorkflow) environmentViewerJSON(view *environmentView, writer io.Writer) Executor {
	return func() error {
		output := common.JSONOutput{}
		output.Values[0].Key = BaseURLKey
		output.Values[0].Value = view.baseURL

		enc := json.NewEncoder(writer)
		return enc.Encode(&output)
	}
}

func (workflow *environmentWorkflow) environmentViewerSHELL(view *environmentView, writer io.Writer) Executor {
	return func() error {
		output := common.JSONOutput{}
		output.Values[0].Key = BaseURLKey
		output.Values[0].Value = view.baseURL

		for _, val := range output.Values {
			fmt.Fprintf(writer, "%s=%s\n", val.Key, val.Value)
		}

		return nil
	}
}

func (workflow *environmentWorkflow) environmentViewerCLI(view *environmentView, writer io.Writer) Executor {
	return func() error {

		fmt.Fprintf(writer, HeaderValueFormat, Bold(EnvironmentHeader), view.name)
		if view.clusterName != "" {
			fmt.Fprintf(writer, StackFormat, Bold(ClusterStack), view.clusterName, colorizeStackStatus(view.clusterStatus))
		}

		if view.vpcStatus == "" {
			fmt.Fprintf(writer, UnmanagedStackFormat, Bold(VPCStack))
		} else {
			fmt.Fprintf(writer, StackFormat, Bold(VPCStack), view.vpcName, colorizeStackStatus(view.vpcStatus))
			fmt.Fprintf(writer, HeaderValueFormat, Bold(BastionHost), view.bastionHost)
		}
		fmt.Fprintf(writer, HeaderValueFormat, Bold(BaseURLHeader), view.baseURL)

		if len(view.instances) > 0 {
			fmt.Fprintf(writer, HeadNewlineHeader, Bold(ContainerInstances))
			printInstanceTable(view.instances, writer)
		}

		fmt.Fprint(writer, NewLine)
		fmt.Fprintf(writer, HeadNewlineHeader, Bold(ServicesHeader))
		printServiceTable(view.services, writer)

		fmt.Fprint(writer, NewLine)

		return nil
	}
}

func printServiceTable(services []*serviceView, writer io.Writer) {
	table := CreateTableSection(writer, ServiceTableHeader)

	for _, service := range services {
		table.Append([]string{
			Bold(service.name),
			service.revision,
			fmt.Sprintf(KeyValueFormat, colorizeStackStatus(service.status), service.statusReason),
			service.lastUpdate,
		})
	}

	table.Render()
}

func buildInstanceViewForECS(containerInstances []common.ContainerInstance, instances []common.Instance) []*instanceView {
	instanceIps := make(map[string]string)
	for _, instance := range instances {
		instanceIps[common.StringValue(instance.InstanceId)] = common.StringValue(instance.PrivateIpAddress)
	}

	instanceViews := make([]*instanceView, len(containerInstances))

	for i, instance := range containerInstances {
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
		instanceViews[i] = &instanceView{
			common.StringValue(instance.Ec2InstanceId),
			instanceType,
			amiID,
			instanceIps[common.StringValue(instance.Ec2InstanceId)],
			availZone,
			common.BoolValue(instance.AgentConnected),
			common.StringValue(instance.Status),
			common.Int64Value(instance.RunningTasksCount),
			cpuAvail,
			memAvail,
		}
	}

	return instanceViews
}

func buildServiceViewForECS(namespace string, serviceName string, environmentName string, stackLister common.StackLister) ([]*serviceView, error) {
	stacks, err := stackLister.ListStacks(common.StackTypeService, namespace)
	if err != nil {
		return nil, err
	}

	services := make([]*serviceView, 0)
	for _, stackValues := range stacks {
		if environmentName != "" && stackValues.Tags[EnvTagKey] != environmentName {
			continue
		}
		if serviceName != "" && stackValues.Tags[SvcTagKey] != serviceName {
			continue
		}

		services = append(services, &serviceView{
			stackValues.Tags[SvcTagKey],
			stackValues.Tags["revision"],
			stackValues.Status,
			stackValues.StatusReason,
			stackValues.LastUpdateTime.Local().Format(LastUpdateTime),
		})
	}

	return services, nil
}

func buildInstanceViewForEKS(nodes *unstructured.UnstructuredList) []*instanceView {
	instanceViews := make([]*instanceView, len(nodes.Items))
	for i, node := range nodes.Items {
		var ip string
		addresses := common.MapGetSlice(node.Object, "status", "addresses")
		for _, address := range addresses {
			if common.MapGetString(address, "type") == "InternalIP" {
				ip = common.MapGetString(address, "address")
			}
		}
		instanceViews[i] = &instanceView{
			common.MapGetString(node.Object, "spec", "externalID"),
			"",
			common.MapGetString(node.Object, "status", "nodeInfo", "kernelVersion"),
			ip,
			common.MapGetString(node.Object, "metadata", "labels", "failure-domain.beta.kubernetes.io/zone"),
			true,
			"",
			0,
			0,
			0,
		}
	}

	return instanceViews
}

func printInstanceTable(instances []*instanceView, writer io.Writer) {
	table := CreateTableSection(writer, EnvironmentAMITableHeader)

	for _, instance := range instances {
		table.Append([]string{
			instance.instanceID,
			instance.instanceType,
			instance.amiID,
			instance.instanceIP,
			instance.availZone,
			fmt.Sprintf(BoolStringFormat, instance.ready),
			instance.status,
			fmt.Sprintf(IntStringFormat, instance.taskCount),
			fmt.Sprintf(IntStringFormat, instance.cpuAvail),
			fmt.Sprintf(IntStringFormat, instance.memAvail),
		})
	}

	table.Render()
}
