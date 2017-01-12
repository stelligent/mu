package workflows

import (
	"bytes"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
	"io"
	"testing"
)

func TestEnvironmentFinder(t *testing.T) {
	assert := assert.New(t)

	env1 := common.Environment{
		Name: "foo",
	}
	env2 := common.Environment{
		Name: "bar",
	}
	config := new(common.Config)
	config.Environments = []common.Environment{env1, env2}

	var environment *common.Environment

	environment = nil
	fooErr := environmentFinder(&environment, config, "foo")()
	assert.NotNil(environment)
	assert.Equal("foo", environment.Name)
	assert.Nil(fooErr)

	environment = nil
	barErr := environmentFinder(&environment, config, "bar")()
	assert.NotNil(environment)
	assert.Equal("bar", environment.Name)
	assert.Nil(barErr)

	environment = nil
	bazErr := environmentFinder(&environment, config, "baz")()
	assert.Nil(environment)
	assert.NotNil(bazErr)
}

func TestNewEnvironmentUpserter(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	upserter := NewEnvironmentUpserter(ctx, "foo")
	assert.NotNil(upserter)
}

type mockedStackManager struct {
	mock.Mock
}

func (m *mockedStackManager) AwaitFinalStatus(stackName string) string {
	args := m.Called(stackName)
	return args.String(0)
}
func (m *mockedStackManager) UpsertStack(stackName string, templateBodyReader io.Reader, stackParameters map[string]string) error {
	args := m.Called(stackName)
	return args.Error(0)
}

func TestEnvironmentVpcUpserter(t *testing.T) {
	assert := assert.New(t)

	environment := &common.Environment{
		Name: "foo",
	}

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManager)
	stackManager.On("AwaitFinalStatus", "mu-vpc-foo").Return(cloudformation.StackStatusCreateComplete)
	stackManager.On("UpsertStack", "mu-vpc-foo").Return(nil)

	err := environmentVpcUpserter(environment, vpcInputParams, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("mu-vpc-foo-VpcId", vpcInputParams["VpcId"])
	assert.Equal("mu-vpc-foo-PublicSubnetAZ1Id", vpcInputParams["PublicSubnetAZ1Id"])
	assert.Equal("mu-vpc-foo-PublicSubnetAZ2Id", vpcInputParams["PublicSubnetAZ2Id"])
	assert.Equal("mu-vpc-foo-PublicSubnetAZ3Id", vpcInputParams["PublicSubnetAZ3Id"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestEnvironmentVpcUpserter_Unmanaged(t *testing.T) {
	assert := assert.New(t)
	yamlConfig :=
		`
---
environments:
  - name: dev
    vpcTarget:
      vpcId: myVpcId
      publicSubnetIds:
        - mySubnetId1
        - mySubnetId2
`
	config, err := loadYamlConfig(yamlConfig)
	assert.Nil(err)

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManager)

	err = environmentVpcUpserter(&config.Environments[0], vpcInputParams, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("myVpcId", vpcInputParams["VpcId"])
	assert.Equal("mySubnetId1", vpcInputParams["PublicSubnetAZ1Id"])
	assert.Equal("mySubnetId2", vpcInputParams["PublicSubnetAZ2Id"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 0)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 0)
}

func loadYamlConfig(yamlString string) (*common.Config, error) {
	config := new(common.Config)
	yamlBuffer := new(bytes.Buffer)
	yamlBuffer.ReadFrom(bytes.NewBufferString(yamlString))
	err := yaml.Unmarshal(yamlBuffer.Bytes(), config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
