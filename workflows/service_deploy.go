package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewServiceDeployer create a new workflow for deploying a service in an environment
func NewServiceDeployer(ctx *common.Context, environmentName string, tag string) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceLoader(&ctx.Config, tag),
		workflow.serviceDeployer(environmentName),
	)
}

func (workflow *serviceWorkflow) serviceDeployer(environmentName string) Executor {
	return func() error {
		log.Debugf("Deploying service '%s' image '%s' to environment '%s'", workflow.serviceName, workflow.serviceImage, environmentName)
		return nil
	}
}
