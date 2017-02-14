package templates

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestNewTemplate(t *testing.T) {
	assert := assert.New(t)

	environment := new(common.Environment)

	templates := []string{"cluster.yml", "vpc.yml"}
	for _, templateName := range templates {
		templateBodyReader, err := NewTemplate(templateName, environment, nil)

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

	environment := new(common.Environment)

	templateBodyReader, err := NewTemplate("invalid-template-name.yml", environment, nil)
	assert.Nil(templateBodyReader)
	assert.NotNil(err)
}

func TestNewTemplate_withOverrides(t *testing.T) {
	assert := assert.New(t)

	overridesYaml :=
		`
---
Resources:
  Foo:
    Type: AWS::S3::Bucket
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: overrideBucketName
`

	overrides := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(overridesYaml), overrides)
	assert.Nil(err)

	templateBodyReader, err := NewTemplate("bucket.yml", nil, overrides)
	assert.Nil(err)
	assert.NotNil(templateBodyReader)

	templateBodyBytes := new(bytes.Buffer)
	templateBodyBytes.ReadFrom(templateBodyReader)
	templateBody := aws.String(templateBodyBytes.String())

	assert.NotNil(templateBody)
	assert.NotEmpty(templateBody)

	fmt.Println(templateBody)
}
