package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
)

// NewPipelineUpserter create a new workflow for upserting a pipeline
func NewPipelineUpserter(ctx *common.Context, tokenProvider func(bool) string) Executor {

	workflow := new(pipelineWorkflow)

	return newWorkflow(
		workflow.serviceFinder(ctx),
		workflow.pipelineBucket(ctx.Config.Region),
		workflow.pipelineUpserter(tokenProvider, ctx.StackManager, ctx.StackManager),
	)
}

// Find the service in config
func (workflow *pipelineWorkflow) serviceFinder(ctx *common.Context) Executor {

	return func() error {
		// Repo Name
		if ctx.Config.Service.Name == "" {
			workflow.serviceName = ctx.Repo.Name
		} else {
			workflow.serviceName = ctx.Config.Service.Name
		}
		return nil
	}
}

// Setup the artifact bucket
func (workflow *pipelineWorkflow) pipelineBucket(region string) Executor {

	return func() error {
		// TODO: determine accountid
		accountID := "324320755747"
		workflow.pipelineBucketName = fmt.Sprintf("codepipeline-%s-%s", region, accountID)

		// TODO: ensure bucket exists

		return nil
	}
}

func (workflow *pipelineWorkflow) pipelineUpserter(tokenProvider func(bool) string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		pipelineStackName := common.CreateStackName(common.StackTypePipeline, workflow.serviceName)
		pipelineStack := stackWaiter.AwaitFinalStatus(pipelineStackName)

		// no target VPC, we need to create/update the VPC stack
		log.Noticef("Upserting Pipeline for service'%s' ...", workflow.serviceName)
		template, err := templates.NewTemplate("pipeline.yml", nil)
		if err != nil {
			return err
		}
		pipelineParams := make(map[string]string)
		pipelineParams["CodePipelineBucket"] = workflow.pipelineBucketName
		// TODO: replace with real values
		pipelineParams["GitHubUser"] = "stelligent"
		pipelineParams["GitHubRepo"] = "microservice-exemplar"
		pipelineParams["GitHubBranch"] = "mu"
		// TODO: Don't set if not provided...allow for UsePrevious on upsert
		pipelineParams["GitHubToken"] = tokenProvider(pipelineStack == nil)
		// TODO: add params for build attributes
		//pipelineParams["BuildType"] = ""
		//pipelineParams["BuildComputeType"] = ""
		//pipelineParams["BuildImage"] = ""

		err = stackUpserter.UpsertStack(pipelineStackName, template, pipelineParams, buildPipelineTags(workflow.serviceName, common.StackTypePipeline))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", pipelineStackName)
		stackWaiter.AwaitFinalStatus(pipelineStackName)

		return nil
	}
}

func buildPipelineTags(serviceName string, stackType common.StackType) map[string]string {
	return map[string]string{
		"type":    string(stackType),
		"service": serviceName,
	}
}
