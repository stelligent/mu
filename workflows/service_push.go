package workflows

import (
	"github.com/stelligent/mu/common"
	"io"
)

// NewServicePusher create a new workflow for pushing a service to a repo
func NewServicePusher(ctx *common.Context, tag string, dockerWriter io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceLoader(ctx, tag),
		workflow.serviceRepoUpserter(&ctx.Config.Service, ctx.StackManager, ctx.StackManager),
		workflow.serviceBuilder(ctx.DockerManager, &ctx.Config, dockerWriter),
		workflow.serviceRegistryAuthenticator(ctx.ClusterManager),
		workflow.servicePusher(ctx.DockerManager, dockerWriter),
	)
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
