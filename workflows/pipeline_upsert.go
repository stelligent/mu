package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
)

// NewPipelineUpserter create a new workflow for upserting a pipeline
func NewPipelineUpserter(ctx *common.Context, tokenProvider func(bool) string) Executor {

	workflow := new(pipelineWorkflow)

	return newWorkflow(
		workflow.serviceFinder("", ctx),
		workflow.pipelineBucket(ctx.StackManager, ctx.StackManager),
		workflow.pipelineUpserter(tokenProvider, ctx.StackManager, ctx.StackManager),
	)
}

// Setup the artifact bucket
func (workflow *pipelineWorkflow) pipelineBucket(stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	return func() error {
		bucketStackName := common.CreateStackName(common.StackTypeBucket, "codepipeline")
		overrides := common.GetStackOverrides(bucketStackName)
		template, err := templates.NewTemplate("bucket.yml", nil, overrides)
		if err != nil {
			return err
		}
		log.Noticef("Upserting Bucket for CodePipeline")
		bucketParams := make(map[string]string)
		bucketParams["BucketPrefix"] = "codepipeline"
		err = stackUpserter.UpsertStack(bucketStackName, template, bucketParams, buildPipelineTags(workflow.serviceName, common.StackTypeBucket))
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
		overrides := common.GetStackOverrides(pipelineStackName)

		log.Noticef("Upserting Pipeline for service '%s' ...", workflow.serviceName)
		template, err := templates.NewTemplate("pipeline.yml", nil, overrides)
		if err != nil {
			return err
		}
		pipelineParams := make(map[string]string)

		sourceRepo := strings.Split(workflow.pipelineConfig.Source.Repo, "/")
		if sourceRepo == nil || len(sourceRepo) != 2 {
			return fmt.Errorf("Invalid source repo %v", workflow.pipelineConfig.Source.Repo)
		}
		pipelineParams["GitHubUser"] = sourceRepo[0]
		pipelineParams["GitHubRepo"] = sourceRepo[1]
		pipelineParams["GitHubToken"] = tokenProvider(pipelineStack == nil)

		if workflow.pipelineConfig.Source.Branch != "" {
			pipelineParams["GitHubBranch"] = workflow.pipelineConfig.Source.Branch
		}

		if workflow.pipelineConfig.Build.Type != "" {
			pipelineParams["BuildType"] = workflow.pipelineConfig.Build.Type
		}
		if workflow.pipelineConfig.Build.ComputeType != "" {
			pipelineParams["BuildComputeType"] = workflow.pipelineConfig.Build.ComputeType
		}

		if workflow.pipelineConfig.Build.Image != "" {
			pipelineParams["BuildImage"] = workflow.pipelineConfig.Build.Image
		}

		if workflow.pipelineConfig.Acceptance.Type != "" {
			pipelineParams["TestType"] = workflow.pipelineConfig.Acceptance.Type
		}
		if workflow.pipelineConfig.Acceptance.ComputeType != "" {
			pipelineParams["TestComputeType"] = workflow.pipelineConfig.Acceptance.ComputeType
		}

		if workflow.pipelineConfig.Acceptance.Image != "" {
			pipelineParams["TestImage"] = workflow.pipelineConfig.Acceptance.Image
		}

		if workflow.pipelineConfig.Acceptance.Environment != "" {
			pipelineParams["TestEnv"] = workflow.pipelineConfig.Acceptance.Environment
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
