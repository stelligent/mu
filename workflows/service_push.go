package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"io"
	"fmt"
)

// NewServicePusher create a new workflow for pushing a service to a repo
func NewServicePusher(ctx *common.Context, tag string, dockerWriter io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceLoader(&ctx.Config, tag),
		workflow.serviceRepoUpserter(ctx.StackManager, ctx.StackManager),
		workflow.serviceBuilder(ctx.DockerManager, &ctx.Config, dockerWriter),
		workflow.serviceRegistryAuthenticator(ctx.ClusterManager),
		workflow.servicePusher(ctx.DockerManager, dockerWriter),
	)
}

func (workflow *serviceWorkflow) serviceRepoUpserter(stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Upsert repo for service '%s'", workflow.serviceName)

		template, err := templates.NewTemplate("repo.yml", nil)
		if err != nil {
			return err
		}

		stackParams := make(map[string]string)
		stackParams["RepoName"] = workflow.serviceName

		ecrStackName := common.CreateStackName(common.StackTypeRepo, workflow.serviceName)

		err = stackUpserter.UpsertStack(ecrStackName, template, stackParams, buildEnvironmentTags(workflow.serviceName, common.StackTypeRepo))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", ecrStackName)
		stack := stackWaiter.AwaitFinalStatus(ecrStackName)
		workflow.serviceImage = fmt.Sprintf("%s:%s", stack.Outputs["RepoUrl"],workflow.serviceTag)
		return nil
	}
}
func (workflow *serviceWorkflow) serviceBuilder(imageBuilder common.DockerImageBuilder, config *common.Config, dockerWriter io.Writer) Executor {
	return func() error {
		log.Noticef("Building service:'%s' as image:%s'", workflow.serviceName, workflow.serviceImage)
		return imageBuilder.ImageBuild(config.Basedir, config.Service.Dockerfile, []string{workflow.serviceImage}, dockerWriter)
	}
}
func (workflow *serviceWorkflow) servicePusher(imagePusher common.DockerImagePusher, dockerWriter io.Writer) Executor {
	return func() error {
		log.Noticef("Pushing service '%s' to '%s'", workflow.serviceName, workflow.serviceImage)
		return imagePusher.ImagePush(workflow.serviceImage, workflow.registryAuth, dockerWriter)
	}
}
