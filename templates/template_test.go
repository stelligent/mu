package templates

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/gobuffalo/packr"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewTemplate(t *testing.T) {
	assert := assert.New(t)

	templates := []string{common.TemplatePolicyDefault, common.TemplatePolicyAllowAll,
		common.TemplateApp, common.TemplateBucket, common.TemplateBuildspec,
		common.TemplateCommonIAM, common.TemplateDatabase, common.TemplateELB,
		common.TemplateEnvEC2, common.TemplateEnvECS, common.TemplatePipelineIAM,
		common.TemplatePipeline, common.TemplateRepo, common.TemplateSchedule,
		common.TemplateServiceEC2, common.TemplateServiceECS, common.TemplateServiceIAM,
		common.TemplateVPCTarget, common.TemplateVPC}

	for _, templateName := range templates {
		templateBody, err := GetAsset(templateName)

		assert.Nil(err)
		assert.NotNil(templateBody)
		assert.NotEmpty(templateBody)
	}
}

func TestNewTemplate_invalid(t *testing.T) {
	assert := assert.New(t)

	templateBodyReader, err := GetAsset("invalid-template-name.yml")
	assert.Empty(templateBodyReader)
	assert.NotNil(err)
}

func TestNewTemplate_assets(t *testing.T) {
	assert := assert.New(t)

	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}
	sess, err := session.NewSessionWithOptions(sessOptions)
	assert.Nil(err)

	svc := cloudformation.New(sess)

	box := packr.NewBox("./assets")
	templates := box.List()
	assert.NotZero(len(templates))
	for _, templateName := range templates {
		if !strings.HasPrefix(templateName, "cloudformation/") {
			continue
		}

		var templateData interface{}
		if templateName == "cloudformation/artifact-pipeline.yml" {
			tdMap := make(map[string]string)
			tdMap["SourceProvider"] = "GitHub"
			tdMap["EnableAcptStage"] = "true"
			tdMap["EnableProdStage"] = "true"
			templateData = tdMap
		}

		templateBody, err := GetAsset(templateName, ExecuteTemplate(templateData))

		if err != nil {
			fmt.Printf("%v", err)
		}
		assert.Nil(err, templateName)
		assert.NotNil(templateBody, templateName)

		if templateBody != "" {

			assert.NotEmpty(templateBody, templateName)

			params := &cloudformation.ValidateTemplateInput{
				TemplateBody: aws.String(templateBody),
			}

			_, err := svc.ValidateTemplate(params)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					if awsErr.Code() == "InvalidClientTokenId" && awsErr.Message() == "The security token included in the request is invalid." {
						t.Skip("Invalid AWS client token id to run CFN template validation")
					}
					if awsErr.Code() == "NoCredentialProviders" {
						t.Skip("No valid AWS credential provider to run CFN template validation")
					}
					if awsErr.Code() == "MissingRegion" {
						t.Skip("No valid AWS region to run CFN template validation")
					}
					if awsErr.Code() == "ValidationError" && awsErr.Message() == "Template format error: Unrecognized resource types: [AWS::EKS::Cluster]" {
						t.Skip("AWS::EKS::Cluster is not recognized by CloudFormation ValidateTemplate yet")
					}
					assert.Fail(awsErr.Code(), awsErr.Message(), templateName)
				}
				assert.Fail(err.Error(), templateName)
			}
		}
	}
}
