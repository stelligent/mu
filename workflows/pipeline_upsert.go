package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
)

// NewPipelineUpserter create a new workflow for upserting a pipeline
func NewPipelineUpserter(ctx *common.Context, tokenProvider func(bool) string) Executor {

	workflow := new(pipelineWorkflow)

	return newWorkflow(
		workflow.serviceFinder(ctx),
		workflow.pipelineBucket(ctx.StackManager, ctx.StackManager),
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

		workflow.pipelineConfig = &ctx.Config.Service.Pipeline
		return nil
	}
}

// Setup the artifact bucket
func (workflow *pipelineWorkflow) pipelineBucket(stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	return func() error {
		template, err := templates.NewTemplate("bucket.yml", nil)
		if err != nil {
			return err
		}
		bucketStackName := common.CreateStackName(common.StackTypeBucket, "codepipeline")
		bucketParams := make(map[string]string)
		bucketParams["BucketPrefix"] = "codepipeline"
		err = stackUpserter.UpsertStack(bucketStackName, template, bucketParams, buildPipelineTags(workflow.serviceName, common.StackTypePipeline))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", bucketStackName)
		stackWaiter.AwaitFinalStatus(bucketStackName)

		return nil
	}
}

func (workflow *pipelineWorkflow) pipelineUpserter(tokenProvider func(bool) string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		pipelineStackName := common.CreateStackName(common.StackTypePipeline, workflow.serviceName)
		pipelineStack := stackWaiter.AwaitFinalStatus(pipelineStackName)

		log.Noticef("Upserting Pipeline for service'%s' ...", workflow.serviceName)
		template, err := templates.NewTemplate("pipeline.yml", nil)
		if err != nil {
			return err
		}
		pipelineParams := make(map[string]string)

		sourceRepo := strings.Split(workflow.pipelineConfig.SourceRepo, "/")
		pipelineParams["GitHubUser"] = sourceRepo[0]
		pipelineParams["GitHubRepo"] = sourceRepo[1]
		pipelineParams["GitHubBranch"] = workflow.pipelineConfig.SourceBranch
		pipelineParams["GitHubToken"] = tokenProvider(pipelineStack == nil)

		pipelineParams["BuildType"] = workflow.pipelineConfig.BuildType
		pipelineParams["BuildComputeType"] = workflow.pipelineConfig.BuildComputeType

		if workflow.pipelineConfig.BuildImage != "" {
			pipelineParams["BuildImage"] = workflow.pipelineConfig.BuildImage
		}
		if workflow.pipelineConfig.BuildImageVersion != "" {
			pipelineParams["BuildImageVersion"] = workflow.pipelineConfig.BuildImageVersion
		}

		if workflow.pipelineConfig.MuBaseurl != "" {
			pipelineParams["MuDownloadBaseurl"] = workflow.pipelineConfig.MuBaseurl
		}

		version := workflow.pipelineConfig.MuVersion
		if version == "" {
			version = common.GetVersion()
			if version == "0.0.0-local" {
				version = ""
			}
		}
		if version != "" {
			pipelineParams["MuDownloadVersion"] = version
		}

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
