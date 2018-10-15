package workflows

import (
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
	workflow.repoName = ctx.Config.Repo.Slug

	if ctx.Config.Repo.Branch != "" {
		workflow.codeBranch = ctx.Config.Repo.Branch
	} else {
		workflow.codeBranch = ctx.Config.Service.Pipeline.Source.Branch
	}

	stackParams := make(map[string]string)

	return newPipelineExecutor(
		workflow.serviceFinder("", ctx),
		workflow.pipelineToken(ctx.Config.Namespace, tokenProvider, ctx.StackManager, stackParams),
		workflow.pipelineBucket(ctx.Config.Namespace, stackParams, ctx.StackManager, ctx.StackManager),
		workflow.codedeployBucket(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
		workflow.pipelineRolesetUpserter(ctx.RolesetManager, ctx.RolesetManager, stackParams),
		workflow.pipelineUpserter(ctx.Config.Namespace, ctx.StackManager, ctx.StackManager, stackParams),
		workflow.pipelineNotifyUpserter(ctx.Config.Namespace, &ctx.Config.Service.Pipeline, ctx.SubscriptionManager))

}

func (workflow *pipelineWorkflow) codedeployBucket(namespace string, service *common.Service, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {

		if service.Pipeline.Build.Bucket != "" {
			workflow.codeDeployBucket = service.Pipeline.Build.Bucket
		} else {
			bucketStackName := common.CreateStackName(namespace, common.StackTypeBucket, "codedeploy")
			log.Noticef("Upserting Bucket for CodeDeploy")
			bucketParams := make(map[string]string)
			bucketParams["Namespace"] = namespace
			bucketParams["BucketPrefix"] = "codedeploy"

			tags := createTagMap(&PipelineTags{
				Type: common.StackTypeBucket,
			})

			err := stackUpserter.UpsertStack(bucketStackName, common.TemplateBucket, nil, bucketParams, tags, "", "")
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

			workflow.codeDeployBucket = stack.Outputs["Bucket"]
		}

		return nil
	}
}

// Setup the artifact bucket
func (workflow *pipelineWorkflow) pipelineBucket(namespace string, params map[string]string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	return func() error {
		if workflow.pipelineConfig.Bucket != "" {
			params["PipelineBucket"] = workflow.pipelineConfig.Bucket
		} else {
			bucketStackName := common.CreateStackName(namespace, common.StackTypeBucket, "codepipeline")
			log.Noticef("Upserting Bucket for CodePipeline")
			bucketParams := make(map[string]string)
			bucketParams["Namespace"] = namespace
			bucketParams["BucketPrefix"] = "codepipeline"

			tags := createTagMap(&PipelineTags{
				Type: common.StackTypeBucket,
			})

			err := stackUpserter.UpsertStack(bucketStackName, common.TemplateBucket, nil, bucketParams, tags, "", "")
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

			params["PipelineBucket"] = stack.Outputs["Bucket"]
		}

		return nil
	}
}

// Fetch token if needed
func (workflow *pipelineWorkflow) pipelineToken(namespace string, tokenProvider func(bool) string, stackWaiter common.StackWaiter, params map[string]string) Executor {
	return func() error {
		pipelineStackName := common.CreateStackName(namespace, common.StackTypePipeline, workflow.serviceName)
		pipelineStack := stackWaiter.AwaitFinalStatus(pipelineStackName)
		if workflow.pipelineConfig.Source.Provider == "GitHub" {
			params["GitHubToken"] = tokenProvider(pipelineStack == nil)
		}
		return nil
	}
}

func (workflow *pipelineWorkflow) pipelineRolesetUpserter(rolesetUpserter common.RolesetUpserter, rolesetGetter common.RolesetGetter, params map[string]string) Executor {
	return func() error {
		environments := make([]string, 0)

		if !workflow.pipelineConfig.Acceptance.Disabled {
			envName := workflow.pipelineConfig.Acceptance.Environment
			if envName == "" {
				environments = append(environments, "acceptance")
			} else {
				environments = append(environments, envName)
			}
		}

		if !workflow.pipelineConfig.Production.Disabled {
			envName := workflow.pipelineConfig.Production.Environment
			if envName == "" {
				environments = append(environments, "production")
			} else {
				environments = append(environments, envName)
			}
		}

		rolesetExecutors := make([]Executor, 0)

		// add executors for environment and service rolesets
		for i := range environments {
			envName := environments[i]
			rolesetExecutors = append(rolesetExecutors, func() error {
				return rolesetUpserter.UpsertEnvironmentRoleset(envName)
			})

			rolesetExecutors = append(rolesetExecutors, func() error {
				return rolesetUpserter.UpsertServiceRoleset(envName, workflow.serviceName, workflow.codeDeployBucket, workflow.databaseName)
			})
		}

		rolesetExecutors = append(rolesetExecutors, func() error {
			err := rolesetUpserter.UpsertPipelineRoleset(workflow.serviceName, params["PipelineBucket"], workflow.codeDeployBucket)
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
		})

		executor := newPipelineExecutor(
			rolesetUpserter.UpsertCommonRoleset,
			newParallelExecutor(rolesetExecutors...),
		)

		return executor()
	}
}

func (workflow *pipelineWorkflow) pipelineUpserter(namespace string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter, params map[string]string) Executor {
	return func() error {
		pipelineStackName := common.CreateStackName(namespace, common.StackTypePipeline, workflow.serviceName)

		log.Noticef("Upserting Pipeline for service '%s' ...", workflow.serviceName)

		pipelineParams, err := PipelineParams(workflow, namespace, params)
		if err != nil {
			return err
		}

		tags := createTagMap(&PipelineTags{
			Type:     common.StackTypePipeline,
			Service:  workflow.serviceName,
			Revision: workflow.codeRevision,
			Repo:     workflow.repoName,
		})

		err = stackUpserter.UpsertStack(pipelineStackName, common.TemplatePipeline, nil, pipelineParams, tags, "", "")
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

		workflow.notificationArn = stack.Outputs["PipelineNotificationTopicArn"]

		return nil
	}
}

// PipelineParams creates a map of params to send to the CFN pipeline template
func PipelineParams(workflow *pipelineWorkflow, namespace string, params map[string]string) (map[string]string, error) {

	pipelineParams := params

	pipelineParams["Namespace"] = namespace
	pipelineParams["ServiceName"] = workflow.serviceName
	pipelineParams["MuFile"] = workflow.muFile
	pipelineParams["SourceProvider"] = workflow.pipelineConfig.Source.Provider
	pipelineParams["SourceRepo"] = workflow.pipelineConfig.Source.Repo

	common.NewMapElementIfNotEmpty(pipelineParams, "SourceBranch", workflow.codeBranch)

	if workflow.pipelineConfig.Source.Provider == "S3" {
		repoParts := strings.Split(workflow.pipelineConfig.Source.Repo, "/")
		pipelineParams["SourceBucket"] = repoParts[0]
		pipelineParams["SourceObjectKey"] = strings.Join(repoParts[1:], "/")
	}

	common.NewMapElementIfNotEmpty(pipelineParams, "BuildType", string(workflow.pipelineConfig.Build.Type))
	common.NewMapElementIfNotEmpty(pipelineParams, "BuildComputeType", string(workflow.pipelineConfig.Build.ComputeType))
	common.NewMapElementIfNotEmpty(pipelineParams, "BuildImage", workflow.pipelineConfig.Build.Image)
	common.NewMapElementIfNotEmpty(pipelineParams, "PipelineBuildTimeout", workflow.pipelineConfig.Build.BuildTimeout)
	common.NewMapElementIfNotEmpty(pipelineParams, "TestType", string(workflow.pipelineConfig.Acceptance.Type))
	common.NewMapElementIfNotEmpty(pipelineParams, "TestComputeType", string(workflow.pipelineConfig.Acceptance.ComputeType))
	common.NewMapElementIfNotEmpty(pipelineParams, "TestImage", workflow.pipelineConfig.Acceptance.Image)
	common.NewMapElementIfNotEmpty(pipelineParams, "AcptEnv", workflow.pipelineConfig.Acceptance.Environment)
	common.NewMapElementIfNotEmpty(pipelineParams, "PipelineBuildAcceptanceTimeout", workflow.pipelineConfig.Acceptance.BuildTimeout)
	common.NewMapElementIfNotEmpty(pipelineParams, "ProdEnv", workflow.pipelineConfig.Production.Environment)
	common.NewMapElementIfNotEmpty(pipelineParams, "PipelineBuildProductionTimeout", workflow.pipelineConfig.Production.BuildTimeout)
	common.NewMapElementIfNotEmpty(pipelineParams, "MuDownloadBaseurl", workflow.pipelineConfig.MuBaseurl)

	pipelineParams["EnableBuildStage"] = strconv.FormatBool(!workflow.pipelineConfig.Build.Disabled)
	pipelineParams["EnableAcptStage"] = strconv.FormatBool(!workflow.pipelineConfig.Acceptance.Disabled)
	pipelineParams["EnableProdStage"] = strconv.FormatBool(!workflow.pipelineConfig.Production.Disabled)

	// get default buildspec
	buildspec, err := templates.GetAsset(common.TemplateBuildspec,
		templates.ExecuteTemplate(nil))
	if err != nil {
		return nil, err
	}
	newlineRegexp := regexp.MustCompile(`\r?\n`)
	buildspecString := newlineRegexp.ReplaceAllString(buildspec, "\\n")

	params["DefaultBuildspec"] = buildspecString

	version := workflow.pipelineConfig.MuVersion
	if version == "" {
		version = common.GetVersion()
		if version == "0.0.0-local" {
			version = ""
		}
	}
	if version != "" {
		params["MuDownloadVersion"] = version
	}

	return pipelineParams, nil
}

func (workflow *pipelineWorkflow) pipelineNotifyUpserter(namespace string, pipeline *common.Pipeline, subManager common.SubscriptionManager) Executor {
	return func() error {
		if len(workflow.notificationArn) > 0 && len(pipeline.Notify) > 0 {
			log.Noticef("Updating pipeline notifications for service '%s' ...", workflow.serviceName)
			for _, notify := range pipeline.Notify {
				sub, _ := subManager.GetSubscription(workflow.notificationArn, "email", notify)
				if sub == nil {
					log.Infof("  Subscribing '%s' to '%s'", notify, workflow.notificationArn)
					err := subManager.CreateSubscription(workflow.notificationArn, "email", notify)
					if err != nil {
						return err
					}
				}

			}
		}
		return nil
	}
}
