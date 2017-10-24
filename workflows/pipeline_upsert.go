package workflows

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
)

// NewPipelineUpserter create a new workflow for upserting a pipeline
func NewPipelineUpserter(ctx *common.Context, tokenProvider func(bool) string) Executor {

	workflow := new(pipelineWorkflow)
	workflow.codeRevision = ctx.Config.Repo.Revision
	workflow.codeBranch = ctx.Config.Repo.Branch
	workflow.repoName = ctx.Config.Repo.Slug

	stackParams := make(map[string]string)

	return newPipelineExecutor(
		workflow.serviceFinder("", ctx),
		workflow.pipelineBucket(ctx.Config.Namespace, ctx.StackManager, ctx.StackManager),
		workflow.pipelineRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, stackParams),
		workflow.pipelineUpserter(ctx.Config.Namespace, tokenProvider, ctx.StackManager, ctx.StackManager, stackParams),
	)
}

// Setup the artifact bucket
func (workflow *pipelineWorkflow) pipelineBucket(namespace string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	return func() error {
		bucketStackName := common.CreateStackName(namespace, common.StackTypeBucket, "codepipeline")
		overrides := common.GetStackOverrides(bucketStackName)
		template, err := templates.NewTemplate("bucket.yml", nil, overrides)
		if err != nil {
			return err
		}
		log.Noticef("Upserting Bucket for CodePipeline")
		bucketParams := make(map[string]string)
		bucketParams["Namespace"] = namespace
		bucketParams["BucketPrefix"] = "codepipeline"

		var pipeTags TagInterface = &PipelineTags{
			Type: common.StackTypeBucket,
		}
		tags, err := concatTags(workflow.pipelineConfig.Tags, pipeTags)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(bucketStackName, template, bucketParams, tags, "")
		if err != nil {
			// ignore error if stack is in progress already
			if !strings.Contains(err.Error(), "_IN_PROGRESS state and can not be updated") {
				return err
			}
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

func (workflow *pipelineWorkflow) pipelineRolesetUpserter(rolesetUpserter common.RolesetUpserter, rolesetGetter common.RolesetGetter, params map[string]string) Executor {
	return func() error {
		err := rolesetUpserter.UpsertCommonRoleset()
		if err != nil {
			return err
		}

		if !workflow.pipelineConfig.Acceptance.Disabled {
			envName := workflow.pipelineConfig.Acceptance.Environment
			if envName == "" {
				envName = "acceptance"
			}
			err := rolesetUpserter.UpsertEnvironmentRoleset(envName)
			if err != nil {
				return err
			}

			err = rolesetUpserter.UpsertServiceRoleset(envName, workflow.serviceName)
			if err != nil {
				return err
			}

		}

		if !workflow.pipelineConfig.Production.Disabled {
			envName := workflow.pipelineConfig.Production.Environment
			if envName == "" {
				envName = "production"
			}
			err := rolesetUpserter.UpsertEnvironmentRoleset(envName)
			if err != nil {
				return err
			}

			err = rolesetUpserter.UpsertServiceRoleset(envName, workflow.serviceName)
			if err != nil {
				return err
			}
		}

		err = rolesetUpserter.UpsertPipelineRoleset(workflow.serviceName)
		if err != nil {
			return err
		}

		pipelineRoleset, err := rolesetGetter.GetPipelineRoleset(workflow.serviceName)
		if err != nil {
			return err
		}

		for roleType, roleArn := range pipelineRoleset {
			if roleArn != "" {
				params[roleType] = roleArn
			}
		}

		return nil
	}
}

func (workflow *pipelineWorkflow) pipelineUpserter(namespace string, tokenProvider func(bool) string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter, params map[string]string) Executor {
	return func() error {
		pipelineStackName := common.CreateStackName(namespace, common.StackTypePipeline, workflow.serviceName)
		pipelineStack := stackWaiter.AwaitFinalStatus(pipelineStackName)
		overrides := common.GetStackOverrides(pipelineStackName)

		log.Noticef("Upserting Pipeline for service '%s' ...", workflow.serviceName)
		template, err := templates.NewTemplate("pipeline.yml", nil, overrides)
		if err != nil {
			return err
		}
		pipelineParams := params

		pipelineParams["Namespace"] = namespace
		pipelineParams["ServiceName"] = workflow.serviceName
		pipelineParams["MuFile"] = workflow.muFile
		pipelineParams["SourceProvider"] = workflow.pipelineConfig.Source.Provider
		pipelineParams["SourceRepo"] = workflow.pipelineConfig.Source.Repo
		pipelineParams["SourceBranch"] = workflow.codeBranch
		pipelineParams["SourceBucket"] = workflow.pipelineConfig.Source.SourceBucket
		pipelineParams["SourceObjectKey"] = workflow.pipelineConfig.Source.SourceObjectKey
		if workflow.pipelineConfig.Source.Provider == "GitHub" {
			pipelineParams["GitHubToken"] = tokenProvider(pipelineStack == nil)
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
			pipelineParams["AcptEnv"] = workflow.pipelineConfig.Acceptance.Environment
		}

		if workflow.pipelineConfig.Production.Environment != "" {
			pipelineParams["ProdEnv"] = workflow.pipelineConfig.Production.Environment
		}

		if workflow.pipelineConfig.MuBaseurl != "" {
			pipelineParams["MuDownloadBaseurl"] = workflow.pipelineConfig.MuBaseurl
		}

		pipelineParams["EnableBuildStage"] = strconv.FormatBool(!workflow.pipelineConfig.Build.Disabled)
		pipelineParams["EnableAcptStage"] = strconv.FormatBool(!workflow.pipelineConfig.Acceptance.Disabled)
		pipelineParams["EnableProdStage"] = strconv.FormatBool(!workflow.pipelineConfig.Production.Disabled)

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
		var pipeTags TagInterface = &PipelineTags{
			Type:     common.StackTypePipeline,
			Service:  workflow.serviceName,
			Revision: workflow.codeRevision,
			Repo:     workflow.repoName,
		}
		tags, err := concatTags(workflow.pipelineConfig.Tags, pipeTags)
		if err != nil {
			return err
		}

		err = stackUpserter.UpsertStack(pipelineStackName, template, pipelineParams, tags, "")
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
