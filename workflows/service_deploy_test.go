package workflows

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServiceDeployer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	deploye := NewServiceDeployer(ctx, "dev", "foo")
	assert.NotNil(deploye)
}

func TestServiceEnvironmentLoader(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-cluster-dev").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})

	params := make(map[string]string)
	workflow := new(serviceWorkflow)
	err := workflow.serviceEnvironmentLoader("dev", stackManager, params)()
	assert.Nil(err)

	assert.Equal("mu-cluster-dev-VpcId", params["VpcId"])
	assert.Equal("mu-cluster-dev-EcsCluster", params["EcsCluster"])
	assert.Equal("mu-cluster-dev-EcsElbHttpListenerArn", params["EcsElbHttpListenerArn"])
	assert.Equal("mu-cluster-dev-EcsElbHttpsListenerArn", params["EcsElbHttpsListenerArn"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
}

func TestServiceEnvironmentLoader_NotFound(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-cluster-dev").Return(nil)

	params := make(map[string]string)

	workflow := new(serviceWorkflow)
	err := workflow.serviceEnvironmentLoader("dev", stackManager, params)()

	assert.NotNil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
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
