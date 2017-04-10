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

	workflow := new(environmentWorkflow)

	workflow.environment = nil
	fooErr := workflow.environmentFinder(config, "foo")()
	assert.NotNil(workflow.environment)
	assert.Equal("foo", workflow.environment.Name)
	assert.Nil(fooErr)

	workflow.environment = nil
	barErr := workflow.environmentFinder(config, "bar")()
	assert.NotNil(workflow.environment)
	assert.Equal("bar", workflow.environment.Name)
	assert.Nil(barErr)

	workflow.environment = nil
	bazErr := workflow.environmentFinder(config, "baz")()
	assert.Nil(workflow.environment)
	assert.NotNil(bazErr)
}

func TestNewEnvironmentUpserter(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	upserter := NewEnvironmentUpserter(ctx, "foo")
	assert.NotNil(upserter)
}

type mockedStackManagerForUpsert struct {
	mock.Mock
}

func (m *mockedStackManagerForUpsert) AwaitFinalStatus(stackName string) *common.Stack {
	args := m.Called(stackName)
	rtn := args.Get(0)
	if rtn == nil {
		return nil
	}
	return rtn.(*common.Stack)
}
func (m *mockedStackManagerForUpsert) UpsertStack(stackName string, templateBodyReader io.Reader, stackParameters map[string]string, stackTags map[string]string) error {
	args := m.Called(stackName, stackParameters)
	return args.Error(0)
}
func (m *mockedStackManagerForUpsert) FindLatestImageID(pattern string) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestEnvironmentEcsUpserter(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-cluster-foo").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-cluster-foo", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("FindLatestImageID").Return("ami-00000", nil)

	err := workflow.environmentEcsUpserter(vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestEnvironmentConsulUpserter_nilProvider(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)

	err := workflow.environmentConsulUpserter(vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 0)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 0)
}

func TestEnvironmentConsulUpserter_ConsulProvider(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}
	workflow.environment.Discovery.Provider = "consul"

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackResult := &common.Stack{
		Status: cloudformation.StackStatusCreateComplete,
		Outputs: map[string]string{
			"ConsulServerAutoScalingGroup": "test-asg",
			"ConsulRpcClientSecurityGroup": "test-sg",
		},
	}
	stackManager.On("AwaitFinalStatus", "mu-consul-foo").Return(stackResult)
	stackManager.On("UpsertStack", "mu-consul-foo", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("FindLatestImageID").Return("ami-00000", nil)

	err := workflow.environmentConsulUpserter(vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

	assert.Equal("test-asg", vpcInputParams["ConsulServerAutoScalingGroup"])
	assert.Equal("test-sg", vpcInputParams["ConsulRpcClientSecurityGroup"])
}

func TestEnvironmentVpcUpserter(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}
	workflow.environment.Cluster.KeyName = "mykey"

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-vpc-foo").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-vpc-foo", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("FindLatestImageID").Return("ami-00000", nil)

	err := workflow.environmentVpcUpserter(vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("mu-vpc-foo-VpcId", vpcInputParams["VpcId"])
	assert.Equal("mu-vpc-foo-EcsSubnetIds", vpcInputParams["EcsSubnetIds"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
	stackManager.AssertNumberOfCalls(t, "FindLatestImageID", 1)
}

func TestEnvironmentVpcUpserter_NoBastion(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-vpc-foo").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-vpc-foo", mock.AnythingOfType("map[string]string")).Return(nil)

	err := workflow.environmentVpcUpserter(vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("mu-vpc-foo-VpcId", vpcInputParams["VpcId"])
	assert.Equal("mu-vpc-foo-EcsSubnetIds", vpcInputParams["EcsSubnetIds"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
	stackManager.AssertNumberOfCalls(t, "FindLatestImageID", 0)
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
      ecsSubnetIds:
        - mySubnetId1
        - mySubnetId2
`
	config, err := loadYamlConfig(yamlConfig)
	assert.Nil(err)

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("UpsertStack", "mu-target-dev", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("AwaitFinalStatus", "mu-target-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})

	workflow := new(environmentWorkflow)
	workflow.environment = &config.Environments[0]

	err = workflow.environmentVpcUpserter(vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("mu-target-dev-VpcId", vpcInputParams["VpcId"])
	assert.Equal("mu-target-dev-EcsSubnetIds", vpcInputParams["EcsSubnetIds"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
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
