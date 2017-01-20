package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
)

type serviceWorkflow struct {
	service *common.Service
}

// NewServicePusher create a new workflow for pushing a service to a repo
func NewServicePusher(ctx *common.Context, tag string) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceLoader(&ctx.Config),
		workflow.serviceRepoUpserter(ctx.StackManager, ctx.StackManager),
		workflow.serviceBuilder(tag),
		workflow.servicePusher(tag),
	)
}

// NewServiceDeployer create a new workflow for deploying a service in an environment
func NewServiceDeployer(ctx *common.Context, environmentName string, tag string) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceLoader(&ctx.Config),
		workflow.serviceDeployer(environmentName, tag),
	)
}

// Find a service in config, by name and set the reference
func (workflow *serviceWorkflow) serviceLoader(config *common.Config) Executor {
	return func() error {
		workflow.service = &config.Service

		log.Debugf("Working with service '%s' version '%s'", workflow.service.Name, workflow.service.Revision)
		return nil
	}
}

func (workflow *serviceWorkflow) serviceRepoUpserter(stackUpserter common.StackUpserter, stackWaiter common.StackManager) Executor {
	return func() error {
		service := workflow.service
		log.Debugf("Upsert repo for service '%s' version '%s'", service.Name, service.Revision)

		template, err := templates.NewTemplate("repo.yml", service)
		if err != nil {
			return err
		}

		stackParams := make(map[string]string)
		stackParams["RepoName"] = service.Name

		ecrStackName := common.CreateStackName(common.StackTypeRepo, service.Name)

		err = stackUpserter.UpsertStack(ecrStackName, template, stackParams, buildEnvironmentTags(service.Name, common.StackTypeRepo))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", ecrStackName)
		stackWaiter.AwaitFinalStatus(ecrStackName)
		return nil
	}
}
func (workflow *serviceWorkflow) serviceBuilder(tag string) Executor {
	return func() error {
		service := workflow.service
		log.Debugf("Building service '%s' version '%s' tag '%s'", service.Name, service.Revision, tag)
		return nil
	}
}
func (workflow *serviceWorkflow) servicePusher(tag string) Executor {
	return func() error {
		service := workflow.service
		log.Debugf("Pushing service '%s' version '%s' tag '%s'", service.Name, service.Revision, tag)
		return nil
	}
}

func (workflow *serviceWorkflow) serviceDeployer(environmentName string, tag string) Executor {
	return func() error {
		service := workflow.service
		log.Debugf("Deploying service '%s' version '%s' tag '%s' to environment '%s'", service.Name, service.Revision, tag, environmentName)
		return nil
	}
}
