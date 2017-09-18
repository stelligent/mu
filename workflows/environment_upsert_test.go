package workflows

import (
	"bytes"
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
	ctx.Config.Namespace = "mu"
	upserter := NewEnvironmentUpserter(ctx, "foo")
	assert.NotNil(upserter)
}

type mockedStackManagerForUpsert struct {
	mock.Mock
	common.StackManager
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
		Name:     "foo",
		Provider: common.EnvProviderEcs,
	}

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-environment-foo").Return(&common.Stack{Status: common.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-environment-foo", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("FindLatestImageID").Return("ami-00000", nil)

	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	err := workflow.environmentUpserter(ctx, vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestEnvironmentProviderConditionals(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = new(common.Environment)
	workflow.environment.Provider = common.EnvProviderEcs

	assert.True(workflow.isEcsProvider()())
	assert.False(workflow.isEc2Provider()())

	workflow.environment.Provider = common.EnvProviderEc2

	assert.False(workflow.isEcsProvider()())
	assert.True(workflow.isEc2Provider()())
}

func TestEnvironmentElbUpserter(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-loadbalancer-foo").Return(&common.Stack{Status: common.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-loadbalancer-foo", mock.AnythingOfType("map[string]string")).Return(nil)

	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	err := workflow.environmentElbUpserter(ctx, vpcInputParams, vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestEnvironmentConsulConditional(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = new(common.Environment)
	workflow.environment.Discovery.Provider = ""

	assert.False(workflow.isConsulEnabled()())

	workflow.environment.Discovery.Provider = "consul"
	assert.True(workflow.isConsulEnabled()())
}

func TestEnvironmentConsulUpserter_ConsulProvider(t *testing.T) {
	assert := assert.New(t)

	workflow := new(environmentWorkflow)
	workflow.environment = &common.Environment{
		Name: "foo",
	}
	workflow.environment.Discovery.Provider = "consul"

	consulInputParams := make(map[string]string)
	ecsInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackResult := &common.Stack{
		Status: common.StackStatusCreateComplete,
		Outputs: map[string]string{
			"ConsulServerAutoScalingGroup": "test-asg",
			"ConsulRpcClientSecurityGroup": "test-sg",
		},
	}
	stackManager.On("AwaitFinalStatus", "mu-consul-foo").Return(stackResult)
	stackManager.On("UpsertStack", "mu-consul-foo", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("FindLatestImageID").Return("ami-00000", nil)

	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	err := workflow.environmentConsulUpserter(ctx, consulInputParams, ecsInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

	assert.Equal("test-asg", ecsInputParams["ConsulServerAutoScalingGroup"])
	assert.Equal("test-sg", ecsInputParams["ConsulRpcClientSecurityGroup"])
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
	stackManager.On("AwaitFinalStatus", "mu-vpc-foo").Return(&common.Stack{Status: common.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-vpc-foo", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("FindLatestImageID").Return("ami-00000", nil)

	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	err := workflow.environmentVpcUpserter(ctx, vpcInputParams, vpcInputParams, vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("mu-vpc-foo-VpcId", vpcInputParams["VpcId"])
	assert.Equal("mu-vpc-foo-InstanceSubnetIds", vpcInputParams["InstanceSubnetIds"])

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
	stackManager.On("AwaitFinalStatus", "mu-vpc-foo").Return(&common.Stack{Status: common.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-vpc-foo", mock.AnythingOfType("map[string]string")).Return(nil)

	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	err := workflow.environmentVpcUpserter(ctx, vpcInputParams, vpcInputParams, vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("mu-vpc-foo-VpcId", vpcInputParams["VpcId"])
	assert.Equal("mu-vpc-foo-InstanceSubnetIds", vpcInputParams["InstanceSubnetIds"])

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
      instanceSubnetIds:
        - mySubnetId1
        - mySubnetId2
`
	config, err := loadYamlConfig(yamlConfig)
	assert.Nil(err)

	vpcInputParams := make(map[string]string)

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("UpsertStack", "mu-target-dev", mock.AnythingOfType("map[string]string")).Return(nil)
	stackManager.On("AwaitFinalStatus", "mu-target-dev").Return(&common.Stack{Status: common.StackStatusCreateComplete})

	workflow := new(environmentWorkflow)
	workflow.environment = &config.Environments[0]

	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	err = workflow.environmentVpcUpserter(ctx, vpcInputParams, vpcInputParams, vpcInputParams, stackManager, stackManager, stackManager)()
	assert.Nil(err)
	assert.Equal("mu-target-dev-VpcId", vpcInputParams["VpcId"])
	assert.Equal("mu-target-dev-InstanceSubnetIds", vpcInputParams["InstanceSubnetIds"])

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
