package workflows

import (
	"bytes"
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"regexp"
	"strings"
)

// NewPipelineUpserter create a new workflow for upserting a pipeline
func NewPipelineUpserter(ctx *common.Context, tokenProvider func(bool) string) Executor {

	workflow := new(pipelineWorkflow)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.repoName = ctx.Config.Repo.Slug

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
		err = stackUpserter.UpsertStack(bucketStackName, template, bucketParams, buildPipelineTags(workflow.serviceName, common.StackTypeBucket, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", bucketStackName)
		stack := stackWaiter.AwaitFinalStatus(bucketStackName)
		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", bucketStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

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

		pipelineParams["SourceProvider"] = workflow.pipelineConfig.Source.Provider
		pipelineParams["SourceRepo"] = workflow.pipelineConfig.Source.Repo
		if workflow.pipelineConfig.Source.Provider == "GitHub" {
			pipelineParams["GitHubToken"] = tokenProvider(pipelineStack == nil)
		}

		if workflow.pipelineConfig.Source.Branch != "" {
			pipelineParams["SourceBranch"] = workflow.pipelineConfig.Source.Branch
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

		if workflow.pipelineConfig.Production.Environment != "" {
			pipelineParams["ProdEnv"] = workflow.pipelineConfig.Production.Environment
		}

		if workflow.pipelineConfig.MuBaseurl != "" {
			pipelineParams["MuDownloadBaseurl"] = workflow.pipelineConfig.MuBaseurl
		}

		// get default buildspec
		buildspec, err := templates.NewTemplate("buildspec.yml", nil, nil)
		if err != nil {
			return err
		}
		buildspecBytes := new(bytes.Buffer)
		buildspecBytes.ReadFrom(buildspec)
		newlineRegexp := regexp.MustCompile(`\r?\n`)
		buildspecString := newlineRegexp.ReplaceAllString(buildspecBytes.String(), "\\n")
		pipelineParams["DefaultBuildspec"] = buildspecString

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

		err = stackUpserter.UpsertStack(pipelineStackName, template, pipelineParams, buildPipelineTags(workflow.serviceName, common.StackTypePipeline, workflow.codeRevision, workflow.repoName))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", pipelineStackName)
		stack := stackWaiter.AwaitFinalStatus(pipelineStackName)
		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", pipelineStackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}
}

func buildPipelineTags(serviceName string, stackType common.StackType, codeRevision string, repoName string) map[string]string {
	return map[string]string{
		"type":     string(stackType),
		"service":  serviceName,
		"revision": codeRevision,
		"repo":     repoName,
	}
}
