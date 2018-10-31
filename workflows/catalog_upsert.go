package workflows

import (
	"fmt"
	"path"
	"strings"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
)

type catalogWorkflow struct {
	pipelineWorkflow
	catalogBucketName string
	catalogBucketURL  string
	kmsKeyID          string
	principalARN      string
	muVersion         string
}

// NewCatalogUpserter create a new workflow for upserting a catalog
func NewCatalogUpserter(ctx *common.Context) Executor {

	workflow := new(catalogWorkflow)
	workflow.pipelineConfig = &ctx.Config.Service.Pipeline

	if ctx.Config.Service.Pipeline.MuVersion != "" {
		workflow.muVersion = ctx.Config.Service.Pipeline.MuVersion
	} else {
		workflow.muVersion = common.GetVersion()
	}
	if workflow.pipelineConfig.Source.Provider == "" {
		workflow.pipelineConfig.Source.Provider = "GitHub"
	}

	productParams := make(map[string]string)
	pipelineParams := make(map[string]string)

	return newPipelineExecutor(
		workflow.catalogCommonRoleset(pipelineParams, ctx.RolesetManager, ctx.RolesetManager),
		newParallelExecutor(
			workflow.catalogBucket(ctx.Config.Namespace, ctx.StackManager, ctx.StackManager),
			workflow.pipelineBucket(ctx.Config.Namespace, pipelineParams, ctx.StackManager, ctx.StackManager),
			workflow.codedeployBucket(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
		),
		workflow.catalogIAM(ctx.Config.Namespace, productParams, &ctx.Config.Catalog, ctx.StackManager, ctx.StackManager),
		workflow.catalogPortfolio(ctx.Config.Namespace, productParams, ctx.StackManager, ctx.StackManager),
		workflow.catalogProducts(ctx.Config.Namespace, &ctx.Config.Catalog, productParams, ctx.ArtifactManager, ctx.StackManager, ctx.StackManager),
		workflow.catalogParams(&ctx.Config, pipelineParams, ctx.StackManager),
		workflow.catalogProductVersions(ctx.Config.Namespace, &ctx.Config.Catalog, pipelineParams, ctx.ArtifactManager, ctx.ExtensionsManager, ctx.RolesetManager, ctx.StackManager, ctx.CatalogManager),
	)

}
func (workflow *catalogWorkflow) catalogCommonRoleset(pipelineParams map[string]string, rolesetUpserter common.RolesetUpserter, rolesetGetter common.RolesetGetter) Executor {
	return func() error {
		err := rolesetUpserter.UpsertCommonRoleset()
		if err != nil {
			return err
		}

		roleset, err := rolesetGetter.GetCommonRoleset()
		if err != nil {
			return err
		}
		pipelineParams["AcptCloudFormationRoleArn"] = roleset["CloudFormationRoleArn"]
		pipelineParams["ProdCloudFormationRoleArn"] = roleset["CloudFormationRoleArn"]
		return nil
	}
}

// Setup the catalog bucket
func (workflow *catalogWorkflow) catalogBucket(namespace string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	return func() error {
		bucketStackName := common.CreateStackName(namespace, common.StackTypeBucket, "servicecatalog")
		log.Noticef("Upserting Bucket for Service Catalog")
		bucketParams := make(map[string]string)
		bucketParams["Namespace"] = namespace
		bucketParams["BucketPrefix"] = "servicecatalog"

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

		workflow.catalogBucketName = stack.Outputs["Bucket"]
		workflow.catalogBucketURL = stack.Outputs["BucketURL"]

		return nil
	}
}

// Setup the catalog IAM
func (workflow *catalogWorkflow) catalogIAM(namespace string, productParams map[string]string, catalog *common.Catalog, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	return func() error {
		log.Noticef("Upserting Catalog IAM")
		stackParams := make(map[string]string)
		stackParams["Namespace"] = namespace
		stackParams["IAMUserNames"] = strings.Join(catalog.IAMUsers, ",")

		tags := createTagMap(&CatalogTags{
			Type: common.StackTypeIam,
		})

		stackName := common.CreateStackName(namespace, common.StackTypeIam, "portfolio", "common")
		err := stackUpserter.UpsertStack(stackName, common.TemplatePortfolioIAM, nil, stackParams, tags, "", "")
		if err != nil {
			// ignore error if stack is in progress already
			if !strings.Contains(err.Error(), "_IN_PROGRESS state and can not be updated") {
				return err
			}
		}

		log.Debugf("Waiting for stack '%s' to complete", stackName)
		stack := stackWaiter.AwaitFinalStatus(stackName)
		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", stackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		workflow.principalARN = stack.Outputs["CatalogGroupARN"]
		workflow.kmsKeyID = stack.Outputs["KmsKeyId"]
		productParams["CatalogRoleARN"] = stack.Outputs["CatalogRoleARN"]

		return nil
	}
}

// Setup the catalog portfolio
func (workflow *catalogWorkflow) catalogPortfolio(namespace string, params map[string]string, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	return func() error {
		log.Noticef("Upserting Catalog Portfolio")
		stackParams := make(map[string]string)
		stackParams["Namespace"] = namespace
		stackParams["PrincipalARN"] = workflow.principalARN

		tags := createTagMap(&CatalogTags{
			Type: common.StackTypePortfolio,
		})

		stackName := common.CreateStackName(namespace, common.StackTypePortfolio, "common")
		err := stackUpserter.UpsertStack(stackName, common.TemplatePortfolio, nil, stackParams, tags, "", "")
		if err != nil {
			// ignore error if stack is in progress already
			if !strings.Contains(err.Error(), "_IN_PROGRESS state and can not be updated") {
				return err
			}
		}

		log.Debugf("Waiting for stack '%s' to complete", stackName)
		stack := stackWaiter.AwaitFinalStatus(stackName)
		if stack == nil {
			return fmt.Errorf("Unable to create stack %s", stackName)
		}
		if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		params["PortfolioId"] = stack.Outputs["PortfolioId"]

		return nil
	}
}

func (workflow *catalogWorkflow) catalogParams(config *common.Config, pipelineParams map[string]string, stackLister common.StackLister) Executor {
	return func() error {
		pipelineParams["RepoVersion"] = config.Repo.Revision
		pipelineParams["RepoName"] = config.Repo.Name
		pipelineParams["MuVersion"] = workflow.muVersion
		pipelineParams["CatalogBucket"] = workflow.catalogBucketName
		pipelineParams["CodeDeployBucket"] = workflow.pipelineWorkflow.codeDeployBucket
		return nil
	}
}

// Setup the catalog product
func (workflow *catalogWorkflow) catalogProducts(namespace string, catalog *common.Catalog, productParams map[string]string, artifactCreator common.ArtifactCreator, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {

	catExecutors := make([]Executor, len(catalog.Pipelines))

	tags := createTagMap(&CatalogTags{
		Type: common.StackTypeProduct,
	})

	for i, p := range catalog.Pipelines {
		pipeline := p
		catExecutors[i] = func() error {
			log.Noticef("Upserting Catalog Product '%s'", pipeline.Name)

			stackParams := common.MapClone(productParams)
			stackParams["Namespace"] = namespace
			stackParams["ProductName"] = pipeline.Name
			stackParams["ProductDescription"] = pipeline.Description
			stackParams["ProductDefaultVersionName"] = "default"
			stackParams["ProductDefaultVersionURL"] = fmt.Sprintf("%s/%s/%s", workflow.catalogBucketURL, stackParams["ProductDefaultVersionName"], "pipeline.yml")
			stackName := common.CreateStackName(namespace, common.StackTypeProduct, pipeline.Name)

			// generate and upload the dependent templates
			templateNames := []string{
				common.TemplatePipeline,
				common.TemplatePipelineIAM,
				common.TemplateServiceIAM,
			}
			for _, templateName := range templateNames {
				// load the template
				templateBody, err := templates.GetAsset(templateName)
				if err != nil {
					return err
				}

				destURI := fmt.Sprintf("s3://%s/%s/%s", workflow.catalogBucketName, stackParams["ProductDefaultVersionName"], path.Base(templateName))
				err = artifactCreator.CreateArtifact(strings.NewReader(templateBody), destURI, workflow.kmsKeyID)
				if err != nil {
					return err
				}
			}

			err := stackUpserter.UpsertStack(stackName, common.TemplateProduct, nil, stackParams, tags, "", "")
			if err != nil {
				// ignore error if stack is in progress already
				if !strings.Contains(err.Error(), "_IN_PROGRESS state and can not be updated") {
					return err
				}
			}

			log.Debugf("Waiting for stack '%s' to complete", stackName)
			stack := stackWaiter.AwaitFinalStatus(stackName)
			if stack == nil {
				return fmt.Errorf("Unable to create stack %s", stackName)
			}
			if strings.HasSuffix(stack.Status, "ROLLBACK_COMPLETE") || !strings.HasSuffix(stack.Status, "_COMPLETE") {
				return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
			}

			return nil
		}
	}

	return newParallelExecutor(catExecutors...)
}

func (workflow *catalogWorkflow) catalogProductVersions(namespace string, catalog *common.Catalog, pipelineParams map[string]string, artifactCreator common.ArtifactCreator, extensionsManager common.ExtensionsManager, rolesetGetter common.RolesetGetter, stackWaiter common.StackWaiter, catalogUpserter common.CatalogUpserter) Executor {
	return func() error {
		for _, pipeline := range catalog.Pipelines {
			templateName := fmt.Sprintf("artifact-pipeline-%s.yml", pipeline.Name)
			productVersions := make(map[string]string)

			for version, pipelineTemplate := range pipeline.Versions {

				// load the template
				templateData := common.MapClone(pipelineParams)
				templateData["ProductVersion"] = version
				err := PipelineParams(&pipelineTemplate, namespace, pipeline.Name, pipelineTemplate.Source.Branch, "mu.yml", templateData)
				if err != nil {
					return err
				}

				if !pipelineTemplate.Acceptance.Disabled {
					templateData["AcptEnv"] = pipelineTemplate.Acceptance.Environment
					if templateData["AcptEnv"] == "" {
						templateData["AcptEnv"] = "acceptance"
					}
					templateData["AcptEnvProvider"], err = rolesetGetter.GetEnvironmentProvider(templateData["AcptEnv"])
					if err != nil {
						return err
					}
				}
				if !pipelineTemplate.Production.Disabled {
					templateData["ProdEnv"] = pipelineTemplate.Production.Environment
					if templateData["ProdEnv"] == "" {
						templateData["ProdEnv"] = "production"
					}
					templateData["ProdEnvProvider"], err = rolesetGetter.GetEnvironmentProvider(templateData["ProdEnv"])
					if err != nil {
						return err
					}
				}

				err = workflow.uploadCommonTemplates(version, extensionsManager, artifactCreator)

				// generate the templates for this pipeline version
				templateBody, err := templates.GetAsset(common.TemplateArtifactPipeline, templates.ExecuteTemplate(templateData),
					templates.DecorateTemplate(extensionsManager, ""))
				if err != nil {
					return err
				}

				artifactURI := fmt.Sprintf("s3://%s/%s/%s", workflow.catalogBucketName, version, templateName)
				err = artifactCreator.CreateArtifact(strings.NewReader(templateBody), artifactURI, workflow.kmsKeyID)
				if err != nil {
					return err
				}

				productVersions[version] = fmt.Sprintf("%s/%s/%s", workflow.catalogBucketURL, version, templateName)
			}

			stackName := common.CreateStackName(namespace, common.StackTypeProduct, pipeline.Name)
			stack := stackWaiter.AwaitFinalStatus(stackName)
			if stack == nil {
				return fmt.Errorf("Unable to find product id for stack '%s'", stackName)
			}
			productID := stack.Outputs["ProductId"]

			err := catalogUpserter.SetProductVersions(productID, productVersions)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (workflow *catalogWorkflow) uploadCommonTemplates(version string, extensionsManager common.ExtensionsManager, artifactCreator common.ArtifactCreator) error {
	// generate and upload the dependent templates
	templateNames := []string{
		common.TemplatePipeline,
		common.TemplatePipelineIAM,
		common.TemplateServiceIAM,
	}
	for _, tn := range templateNames {
		// load the template
		templateData := make(map[string]string)
		templateBody, err := templates.GetAsset(tn, templates.ExecuteTemplate(templateData),
			templates.DecorateTemplate(extensionsManager, ""))
		if err != nil {
			return err
		}

		destURI := fmt.Sprintf("s3://%s/%s/%s", workflow.catalogBucketName, version, path.Base(tn))
		err = artifactCreator.CreateArtifact(strings.NewReader(templateBody), destURI, workflow.kmsKeyID)
		if err != nil {
			return err
		}
	}
	return nil
}
