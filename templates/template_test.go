package templates

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
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

	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	templateBodyBytes := new(bytes.Buffer)
	templateBodyBytes.ReadFrom(templateBodyReader)
	templateBody := templateBodyBytes.String()

	assert.NotNil(templateBody)
	assert.NotEmpty(templateBody)

	finalMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal(templateBodyBytes.Bytes(), finalMap)
	assert.Equal("mu-bucket-${BucketPrefix}", nestedMap(finalMap, "Outputs", "Bucket", "Export", "Name")["Fn::Sub"])
}

func TestTempalate_fixupYaml(t *testing.T) {
	assert := assert.New(t)

	rawYaml :=
		`
---
Resources:
  Foo:
    Type: AWS::S3::Bucket
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Ref BucketName
      OtherValue: !Sub ${SomeStuff}
      LongString: !Sub |
        foo
        bar baz
        bam!

        after empty line
      FinalValue: hi
      ListOfRefs:
      - !Ref Alpha
      - !Ref Beta
      IfValue: !If [ Foo, "Bar", "Baz" ]

      ### Disabling following yaml...unable to handle in Golang
      #AvailabilityZone: !Select [ 1, !GetAZs '']
      #DeepMap:
      #- "Fn::Equals": [!Ref ElbInternal, 'true']
      #- "Fn::Join": [ "", !Ref PathPattern]



`
	fixedYaml := fixupYaml([]byte(rawYaml))

	result := make(map[interface{}]interface{})
	err := yaml.Unmarshal(fixedYaml, result)
	assert.Nil(err)
	assert.Equal("BucketName", nestedMap(result, "Resources", "Bucket", "Properties", "BucketName")["Ref"])
	assert.Equal("${SomeStuff}", nestedMap(result, "Resources", "Bucket", "Properties", "OtherValue")["Fn::Sub"])
	assert.Equal("foo\nbar baz\nbam!\n\nafter empty line\n", nestedMap(result, "Resources", "Bucket", "Properties", "LongString")["Fn::Sub"])
	assert.Equal("hi", nestedMap(result, "Resources", "Bucket", "Properties")["FinalValue"])

	ifVal := nestedMap(result, "Resources", "Bucket", "Properties", "IfValue")["Fn::If"].([]interface{})
	assert.Equal("Foo", ifVal[0])
	assert.Equal("Bar", ifVal[1])
	assert.Equal("Baz", ifVal[2])

	listOfRefs := nestedMap(result, "Resources", "Bucket", "Properties")["ListOfRefs"].([]interface{})
	ref1 := listOfRefs[0].(map[interface{}]interface{})
	ref2 := listOfRefs[1].(map[interface{}]interface{})
	assert.Equal("Alpha", ref1["Ref"])
	assert.Equal("Beta", ref2["Ref"])
	/*
		deepMap := nestedMap(result, "Resources", "Bucket", "Properties")["DeepMap"].([]interface{})
		dm1 := deepMap[0].(map[interface{}]interface{})["Fn::Equals"].([]interface{})
		assert.Equal("ElbInternal", dm1[0].(map[interface{}]interface{})["Ref"])
		assert.Equal("true", dm1[1])
		dm2 := deepMap[1].(map[interface{}]interface{})["Fn::Join"].([]interface{})
		assert.Equal("", dm2[0])
		assert.Equal("PathPattern", dm2[1].(map[interface{}]interface{})["Ref"])
	*/

}

func TestNewTemplate_assets(t *testing.T) {
	assert := assert.New(t)

	overrides := make(map[interface{}]interface{})

	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}
	sess, err := session.NewSessionWithOptions(sessOptions)
	assert.Nil(err)

	svc := cloudformation.New(sess)

	templates := []string{"bucket.yml", "cluster.yml", "pipeline.yml", "repo.yml", "service.yml", "vpc.yml", "vpc-target.yml"}
	for _, templateName := range templates {
		templateBodyReader, err := NewTemplate(templateName, nil, overrides)

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
				assert.Fail(err.Error(), templateName)
			}

		}
	}
}

func nestedMap(m map[interface{}]interface{}, keys ...string) map[interface{}]interface{} {
	rtn := m
	for _, key := range keys {
		rtn = rtn[key].(map[interface{}]interface{})
	}
	return rtn
}
