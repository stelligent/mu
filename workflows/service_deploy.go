package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strconv"
	"strings"
)

// NewServiceDeployer create a new workflow for deploying a service in an environment
func NewServiceDeployer(ctx *common.Context, environmentName string, tag string) Executor {

	workflow := new(serviceWorkflow)
	ecsImportParams := make(map[string]string)

	return newWorkflow(
		workflow.serviceLoader(ctx, tag),
		workflow.serviceRepoUpserter(&ctx.Config.Service, ctx.StackManager, ctx.StackManager),
		workflow.serviceEnvironmentLoader(environmentName, ctx.StackManager, ecsImportParams),
		workflow.serviceDeployer(&ctx.Config.Service, ecsImportParams, environmentName, ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *serviceWorkflow) serviceEnvironmentLoader(environmentName string, stackWaiter common.StackWaiter, ecsImportParams map[string]string) Executor {
	return func() error {
		ecsStackName := common.CreateStackName(common.StackTypeCluster, environmentName)
		ecsStack := stackWaiter.AwaitFinalStatus(ecsStackName)

		if ecsStack == nil {
			return fmt.Errorf("Unable to find stack '%s' for environment '%s'", ecsStackName, environmentName)
		}

		ecsImportParams["VpcId"] = fmt.Sprintf("%s-VpcId", ecsStackName)
		ecsImportParams["EcsCluster"] = fmt.Sprintf("%s-EcsCluster", ecsStackName)
		ecsImportParams["EcsElbHttpListenerArn"] = fmt.Sprintf("%s-EcsElbHttpListenerArn", ecsStackName)
		ecsImportParams["EcsElbHttpsListenerArn"] = fmt.Sprintf("%s-EcsElbHttpsListenerArn", ecsStackName)

		return nil
	}
}

func (workflow *serviceWorkflow) serviceDeployer(service *common.Service, stackParams map[string]string, environmentName string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Deploying service '%s' to '%s' from '%s'", workflow.serviceName, environmentName, workflow.serviceImage)

		stackParams["ServiceName"] = workflow.serviceName
		stackParams["ImageUrl"] = workflow.serviceImage
		if service.Port != 0 {
			stackParams["ServicePort"] = strconv.Itoa(service.Port)
		}
		if service.HealthEndpoint != "" {
			stackParams["ServiceHealthEndpoint"] = service.HealthEndpoint
		}
		if service.CPU != 0 {
			stackParams["ServiceCpu"] = strconv.Itoa(service.CPU)
		}
		if service.Memory != 0 {
			stackParams["ServiceMemory"] = strconv.Itoa(service.Memory)
		}
		if service.DesiredCount != 0 {
			stackParams["ServiceDesiredCount"] = strconv.Itoa(service.DesiredCount)
		}
		if len(service.PathPatterns) > 0 {
			stackParams["PathPattern"] = strings.Join(service.PathPatterns, ",")
		}

		template, err := templates.NewTemplate("service.yml", service)
		if err != nil {
			return err
		}

		svcStackName := common.CreateStackName(common.StackTypeService, workflow.serviceName, environmentName)
		err = stackUpserter.UpsertStack(svcStackName, template, stackParams, buildServiceTags(workflow.serviceName, environmentName, common.StackTypeService))
		if err != nil {
			return err
		}
		log.Debugf("Waiting for stack '%s' to complete", svcStackName)
		stackWaiter.AwaitFinalStatus(svcStackName)

		return nil
	}
}

func buildServiceTags(serviceName string, environmentName string, stackType common.StackType) map[string]string {
	return map[string]string{
		"type":        string(stackType),
		"environment": environmentName,
		"service":     serviceName,
	}
}
