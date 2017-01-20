package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewServiceDeployer create a new workflow for deploying a service in an environment
func NewServiceDeployer(ctx *common.Context, environmentName string, tag string) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceLoader(&ctx.Config),
		workflow.serviceDeployer(environmentName, tag),
	)
}

func (workflow *serviceWorkflow) serviceDeployer(environmentName string, tag string) Executor {
	return func() error {
		service := workflow.service
		log.Debugf("Deploying service '%s' version '%s' tag '%s' to environment '%s'", service.Name, service.Revision, tag, environmentName)
		return nil
	}
}
