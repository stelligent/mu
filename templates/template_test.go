package templates

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/gobuffalo/packr"
	"github.com/stretchr/testify/assert"
)

func TestNewTemplate(t *testing.T) {
	assert := assert.New(t)

	templates := []string{"elb.yml", "vpc.yml"}
	for _, templateName := range templates {
		templateBodyReader, err := NewTemplate(templateName, nil)

		assert.Nil(err)
		assert.NotNil(templateBodyReader)

		templateBodyBytes := new(bytes.Buffer)
		templateBodyBytes.ReadFrom(templateBodyReader)
		templateBody := aws.String(templateBodyBytes.String())

		assert.NotNil(templateBody)
		assert.NotEmpty(templateBody)
	}
}

func TestNewTemplate_invalid(t *testing.T) {
	assert := assert.New(t)

	templateBodyReader, err := NewTemplate("invalid-template-name.yml", nil)
	assert.Nil(templateBodyReader)
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
		if templateName == "buildspec.yml" || strings.HasPrefix(templateName, "policies/") {
			continue
		}

		templateBodyReader, err := NewTemplate(templateName, nil)

		assert.Nil(err, templateName)
		assert.NotNil(templateBodyReader, templateName)

		if templateBodyReader != nil {
			templateBodyBytes := new(bytes.Buffer)
			templateBodyBytes.ReadFrom(templateBodyReader)
			templateBody := aws.String(templateBodyBytes.String())

			assert.NotNil(templateBody, templateName)
			assert.NotEmpty(templateBody, templateName)

			params := &cloudformation.ValidateTemplateInput{
				TemplateBody: templateBody,
			}

			_, err := svc.ValidateTemplate(params)
			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok {
					if awsErr.Code() == "RequestError" && awsErr.Message() == "send request failed" {
						return
					}
					assert.Fail(awsErr.Code(), awsErr.Message(), templateName)
				}
				assert.Fail(err.Error(), templateName)
			}

		}
	}
}

func TestNewPolicy(t *testing.T) {
	assert := assert.New(t)

	templateBodyReader, err := NewPolicy("default.json")

	assert.Nil(err)
	assert.NotNil(templateBodyReader)

	templateBodyBytes := new(bytes.Buffer)
	templateBodyBytes.ReadFrom(templateBodyReader)
	templateBody := aws.String(templateBodyBytes.String())

	assert.NotNil(templateBody)
	assert.NotEmpty(templateBody)
}
