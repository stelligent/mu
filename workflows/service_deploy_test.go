package workflows

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewServiceDeployer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	deploye := NewServiceDeployer(ctx, "dev", "foo")
	assert.NotNil(deploye)
}

type mockedElbManager struct {
	mock.Mock
}

func (m *mockedElbManager) ListRules(listenerArn string) ([]*elbv2.Rule, error) {
	args := m.Called(listenerArn)
	return args.Get(0).([]*elbv2.Rule), nil
}

func TestServiceEnvironmentLoader_Create(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForUpsert)
	outputs := make(map[string]string)
	outputs["EcsElbHttpListenerArn"] = "foo"
	outputs["EcsElbHttpsListenerArn"] = "foo"
	stackManager.On("AwaitFinalStatus", "mu-cluster-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete, Outputs: outputs}).Once()
	stackManager.On("AwaitFinalStatus", "mu-service-myservice-dev").Return(nil).Once()
	stackManager.On("AwaitFinalStatus", "mu-database-myservice-dev").Return(nil).Once()

	paramManager := new(mockedParamManager)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]*elbv2.Rule{
		{Priority: aws.String("15")},
		{Priority: aws.String("5")},
		{Priority: aws.String("10")},
	})

	params := make(map[string]string)
	workflow := new(serviceWorkflow)
	workflow.serviceName = "myservice"
	err := workflow.serviceEnvironmentLoader("dev", stackManager, params, elbRuleLister, paramManager)()
	assert.Nil(err)

	assert.Equal("mu-cluster-dev-VpcId", params["VpcId"])
	assert.Equal("mu-cluster-dev-EcsCluster", params["EcsCluster"])
	assert.Equal("mu-cluster-dev-EcsElbHttpListenerArn", params["EcsElbHttpListenerArn"])
	assert.Equal("mu-cluster-dev-EcsElbHttpsListenerArn", params["EcsElbHttpsListenerArn"])
	assert.Equal("16", params["ListenerRulePriority"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 3)
	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)
}
func TestServiceEnvironmentLoader_Update(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForUpsert)
	outputs := make(map[string]string)
	outputs["EcsElbHttpListenerArn"] = "foo"
	outputs["EcsElbHttpsListenerArn"] = "foo"
	stackManager.On("AwaitFinalStatus", "mu-cluster-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete, Outputs: outputs}).Once()
	stackManager.On("AwaitFinalStatus", "mu-service-myservice-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete, Outputs: outputs}).Once()
	stackManager.On("AwaitFinalStatus", "mu-database-myservice-dev").Return(nil).Once()

	paramManager := new(mockedParamManager)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]*elbv2.Rule{
		{Priority: aws.String("15")},
		{Priority: aws.String("5")},
		{Priority: aws.String("10")},
	})

	params := make(map[string]string)
	workflow := new(serviceWorkflow)
	workflow.serviceName = "myservice"
	err := workflow.serviceEnvironmentLoader("dev", stackManager, params, elbRuleLister, paramManager)()
	assert.Nil(err)

	assert.Equal("", params["ListenerRulePriority"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 3)
	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)
}
func TestServiceEnvironmentLoader_StaticPriority(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForUpsert)
	outputs := make(map[string]string)
	outputs["EcsElbHttpListenerArn"] = "foo"
	outputs["EcsElbHttpsListenerArn"] = "foo"
	stackManager.On("AwaitFinalStatus", "mu-cluster-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete, Outputs: outputs}).Once()
	stackManager.On("AwaitFinalStatus", "mu-service-myservice-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete, Outputs: outputs}).Once()
	stackManager.On("AwaitFinalStatus", "mu-database-myservice-dev").Return(nil).Once()

	paramManager := new(mockedParamManager)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]*elbv2.Rule{
		{Priority: aws.String("15")},
		{Priority: aws.String("5")},
		{Priority: aws.String("10")},
	})

	params := make(map[string]string)
	workflow := new(serviceWorkflow)
	workflow.serviceName = "myservice"
	workflow.priority = 77
	err := workflow.serviceEnvironmentLoader("dev", stackManager, params, elbRuleLister, paramManager)()
	assert.Nil(err)

	assert.Equal("77", params["ListenerRulePriority"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 3)
	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)
}

func TestServiceEnvironmentLoader_NotFound(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-cluster-dev").Return(nil)

	paramManager := new(mockedParamManager)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return(nil)

	params := make(map[string]string)

	workflow := new(serviceWorkflow)
	err := workflow.serviceEnvironmentLoader("dev", stackManager, params, elbRuleLister, paramManager)()

	assert.NotNil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "ListRules", 0)
}

func TestServiceGetMaxPriority(t *testing.T) {
	assert := assert.New(t)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]*elbv2.Rule{
		{Priority: aws.String("15")},
		{Priority: aws.String("5")},
		{Priority: aws.String("10")},
	})

	max := getMaxPriority(elbRuleLister, "foo")

	assert.Equal(15, max)

	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)

}

func TestServiceDeployer(t *testing.T) {
	assert := assert.New(t)

	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-service-foo-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-service-foo-dev").Return(nil)

	config := new(common.Config)
	config.Service.Name = "foo"

	params := make(map[string]string)

	workflow := new(serviceWorkflow)
	workflow.serviceName = "foo"
	err := workflow.serviceDeployer(&config.Service, params, "dev", stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

}
