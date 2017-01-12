package templates

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTemplate(t *testing.T) {
	assert := assert.New(t)

	environment := new(common.Environment)

	templates := []string{"cluster.yml", "vpc.yml"}
	for _, templateName := range templates {
		templateBodyReader, err := NewTemplate(templateName, environment)

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

	templateBodyReader, err := NewTemplate("invalid-template-name.yml", environment)
	assert.Nil(templateBodyReader)
	assert.NotNil(err)
}
