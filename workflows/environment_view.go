package workflows

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/stelligent/mu/common"
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
	view.instances = make([]*instanceView, 0)
	view.services = make([]*serviceView, 0)

	var environmentViewer func() error
	if format == JSON {
		environmentViewer = workflow.environmentViewerJSON(view, writer)
	} else if format == SHELL {
		environmentViewer = workflow.environmentViewerSHELL(view, writer)
	} else {
		environmentViewer = workflow.environmentViewerCLI(view, writer)
	}

	return newPipelineExecutor(
		workflow.environmentLoader(ctx.Config.Namespace, environmentName, ctx.StackManager, view),
		newConditionalExecutor(
			workflow.isKubernetesProvider(),
			newPipelineExecutor(
				workflow.connectKubernetes(ctx.Config.Namespace, ctx.KubernetesResourceManagerProvider),
				workflow.environmentEksIngressLoader(view),
				workflow.environmentEksNodeLoader(&view.instances),
				workflow.environmentEksServiceLoader(&view.services),
			),
			workflow.environmentCFNServiceLoader(ctx.Config.Namespace, environmentName, "", ctx.StackManager, &view.services),
		),
		newConditionalExecutor(
			workflow.isEcsProvider(),
			newPipelineExecutor(
				workflow.environmentEcsInstanceLoader(ctx.Config.Namespace, environmentName, ctx.ClusterManager, ctx.InstanceManager, &view.instances),
			),
			nil,
		),
		environmentViewer,
	)
}

func (workflow *environmentWorkflow) environmentLoader(namespace string, environmentName string, stackGetter common.StackGetter, view *environmentView) Executor {
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

		view.name = environmentName
		view.provider = common.EnvProvider(clusterStack.Tags["provider"])
		view.clusterName = clusterStackName
		view.clusterStatus = clusterStack.Status
		view.vpcName = vpcStackName
		if vpcStack != nil {
			view.vpcStatus = vpcStack.Status
			view.bastionHost = vpcStack.Outputs[BastionHostKey]
		}

		if lbStack != nil {
			view.baseURL = lbStack.Outputs[BaseURLValueKey]
		} else {
			view.baseURL = clusterStack.Outputs[BaseURLValueKey]
		}

		return nil
	}
}

func (workflow *environmentWorkflow) environmentEksIngressLoader(view *environmentView) Executor {
	return func() error {
		ingressList, err := workflow.kubernetesResourceManager.ListResources("v1", "Service", "mu-ingress")
		if err != nil {
			return err
		}
		for _, ingress := range ingressList.Items {
			if common.MapGetString(ingress.Object, "metadata", "name") == "nginx-ingress-service" {
				host := common.MapGetString(ingress.Object, "status", "loadBalancer", "ingress", 0, "hostname")
				proto := common.MapGetString(ingress.Object, "spec", "ports", 0, "name")
				view.baseURL = fmt.Sprintf("%s://%s", proto, host)
			}
		}
		return nil
	}
}
func (workflow *environmentWorkflow) environmentEksNodeLoader(instances *[]*instanceView) Executor {
	return func() error {
		nodes, err := workflow.kubernetesResourceManager.ListResources("v1", "Node", "")
		if err != nil {
			log.Warningf("Unable to list nodes: %v", err)
			return nil
		}

		for _, node := range nodes.Items {
			var ip string
			addresses := common.MapGetSlice(node.Object, "status", "addresses")
			for _, address := range addresses {
				if common.MapGetString(address, "type") == "InternalIP" {
					ip = common.MapGetString(address, "address")
				}
			}
			*instances = append(*instances, &instanceView{
				common.MapGetString(node.Object, "spec", "externalID"),
				"",
				"",
				ip,
				common.MapGetString(node.Object, "metadata", "labels", "failure-domain.beta.kubernetes.io/zone"),
				true,
				"",
				-1,
				-1,
				-1,
			})
		}

		return nil
	}
}
func (workflow *environmentWorkflow) environmentEksServiceLoader(services *[]*serviceView) Executor {
	return func() error {
		namespaces, err := workflow.kubernetesResourceManager.ListResources("v1", "Namespace", "")
		if err != nil {
			return err
		}

		for _, ns := range namespaces.Items {
			nstype := common.MapGetString(ns.Object, "metadata", "annotations", "mu/type")
			if nstype == common.StackTypeService {
				*services = append(*services, &serviceView{
					common.MapGetString(ns.Object, "metadata", "annotations", "mu/service"),
					common.MapGetString(ns.Object, "metadata", "annotations", "mu/revision"),
					"",
					"",
					"",
				})
			}
		}
		return nil
	}
}

func (workflow *environmentWorkflow) environmentEcsInstanceLoader(namespace string, environmentName string, clusterInstanceLister common.ClusterInstanceLister, instanceLister common.InstanceLister, instanceViews *[]*instanceView) Executor {
	return func() error {
		clusterName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
		containerInstances, err := clusterInstanceLister.ListInstances(clusterName)
		if err == nil {
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
			*instanceViews = append(*instanceViews, &instanceView{
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
			})
		}
		return nil
	}
}
func (workflow *environmentWorkflow) environmentCFNServiceLoader(namespace string, environmentName string, serviceName string, stackLister common.StackLister, serviceViews *[]*serviceView) Executor {
	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeService, namespace)
		if err != nil {
			return err
		}

		for _, stackValues := range stacks {
			if environmentName != "" && stackValues.Tags[EnvTagKey] != environmentName {
				continue
			}
			if serviceName != "" && stackValues.Tags[SvcTagKey] != serviceName {
				continue
			}

			*serviceViews = append(*serviceViews, &serviceView{
				stackValues.Tags[SvcTagKey],
				stackValues.Tags["revision"],
				stackValues.Status,
				stackValues.StatusReason,
				stackValues.LastUpdateTime.Local().Format(LastUpdateTime),
			})
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
		fmt.Fprintf(writer, HeaderValueFormat, Bold("Provider"), view.provider)
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
