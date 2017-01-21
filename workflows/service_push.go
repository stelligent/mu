package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"io"
)

// NewServicePusher create a new workflow for pushing a service to a repo
func NewServicePusher(ctx *common.Context, tag string, dockerWriter io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceLoader(&ctx.Config),
		workflow.serviceRepoUpserter(ctx.StackManager, ctx.StackManager),
		workflow.serviceBuilder(ctx.DockerManager, &ctx.Config, tag, dockerWriter),
		workflow.servicePusher(tag),
	)
}

func (workflow *serviceWorkflow) serviceRepoUpserter(stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		service := workflow.service
		log.Noticef("Upsert repo for service '%s'", service.Name)

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
func (workflow *serviceWorkflow) serviceBuilder(imageBuilder common.DockerImageBuilder, config *common.Config, tag string, dockerWriter io.Writer) Executor {
	return func() error {
		service := workflow.service

		if tag == "" {
			tag = service.Revision
		}
		log.Noticef("Building service '%s' tag '%s'", service.Name, tag)

		imageBuilder.ImageBuild(config.Basedir, service.Dockerfile, []string{tag}, dockerWriter)

		return nil
	}
}
func (workflow *serviceWorkflow) servicePusher(tag string) Executor {
	return func() error {
		service := workflow.service
		log.Noticef("Pushing service '%s' tag '%s'", service.Name, tag)
		return nil
	}
}
