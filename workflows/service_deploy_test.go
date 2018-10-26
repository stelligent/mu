package workflows

import (
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewServiceDeployer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	ctx.Config.Namespace = "mu"
	deployer := NewServiceDeployer(ctx, "dev", "foo")
	assert.NotNil(deployer)
}

type mockedElbManager struct {
	mock.Mock
}

func (m *mockedElbManager) ListRules(listenerArn string) ([]common.ElbRule, error) {
	args := m.Called(listenerArn)
	return args.Get(0).([]common.ElbRule), nil
}

func TestServiceApplyCommon_Create(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForUpsert)
	outputs := make(map[string]string)
	outputs["ElbHttpListenerArn"] = "foo"
	outputs["ElbHttpsListenerArn"] = "foo"

	stackManager.On("AwaitFinalStatus", "mu-service-myservice-dev").Return(nil).Once()
	stackManager.On("AwaitFinalStatus", "mu-database-myservice-dev").Return(nil).Once()

	paramManager := new(mockedParamManager)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]common.ElbRule{
		{Priority: stringRef("15")},
		{Priority: stringRef("5")},
		{Priority: stringRef("10")},
	})

	service := new(common.Service)
	params := make(map[string]string)
	workflow := new(serviceWorkflow)
	workflow.serviceName = "myservice"
	workflow.envStack = &common.Stack{Name: "mu-environment-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	workflow.lbStack = &common.Stack{Name: "mu-loadbalancer-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	err := workflow.serviceApplyCommonParams("mu", service, params, "dev", stackManager, elbRuleLister, paramManager)()
	assert.Nil(err)

	assert.Equal("mu-environment-dev-VpcId", params["VpcId"])
	assert.Equal("mu-loadbalancer-dev-ElbHttpListenerArn", params["ElbHttpListenerArn"])
	assert.Equal("mu-loadbalancer-dev-ElbHttpsListenerArn", params["ElbHttpsListenerArn"])
	assert.Equal("16", params["PathListenerRulePriority"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)
}

func TestServiceApplyCommon_Update(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForUpsert)
	outputs := make(map[string]string)
	outputs["ElbHttpListenerArn"] = "foo"
	outputs["ElbHttpsListenerArn"] = "foo"
	stackManager.On("AwaitFinalStatus", "mu-service-myservice-dev").Return(&common.Stack{Status: common.StackStatusCreateComplete, Outputs: outputs}).Once()
	stackManager.On("AwaitFinalStatus", "mu-database-myservice-dev").Return(nil).Once()

	paramManager := new(mockedParamManager)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]common.ElbRule{
		{Priority: stringRef("15")},
		{Priority: stringRef("5")},
		{Priority: stringRef("10")},
	})

	service := new(common.Service)
	params := make(map[string]string)
	workflow := new(serviceWorkflow)
	workflow.serviceName = "myservice"
	workflow.envStack = &common.Stack{Name: "mu-environment-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	workflow.lbStack = &common.Stack{Name: "mu-loadbalancer-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	err := workflow.serviceApplyCommonParams("mu", service, params, "dev", stackManager, elbRuleLister, paramManager)()
	assert.Nil(err)

	assert.Equal("", params["ListenerRulePriority"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)
}
func TestServiceApplyCommon_StaticPriority(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForUpsert)
	outputs := make(map[string]string)
	outputs["ElbHttpListenerArn"] = "foo"
	outputs["ElbHttpsListenerArn"] = "foo"
	stackManager.On("AwaitFinalStatus", "mu-service-myservice-dev").Return(&common.Stack{Status: common.StackStatusCreateComplete, Outputs: outputs}).Once()
	stackManager.On("AwaitFinalStatus", "mu-database-myservice-dev").Return(nil).Once()

	paramManager := new(mockedParamManager)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]common.ElbRule{
		{Priority: stringRef("15")},
		{Priority: stringRef("5")},
		{Priority: stringRef("10")},
	})

	service := new(common.Service)
	params := make(map[string]string)
	workflow := new(serviceWorkflow)
	workflow.serviceName = "myservice"
	workflow.envStack = &common.Stack{Name: "mu-environment-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	workflow.lbStack = &common.Stack{Name: "mu-loadbalancer-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	workflow.priority = 77
	err := workflow.serviceApplyCommonParams("mu", service, params, "dev", stackManager, elbRuleLister, paramManager)()
	assert.Nil(err)

	assert.Equal("77", params["PathListenerRulePriority"])

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)
}

func TestServiceEnvLoader_NotFound(t *testing.T) {
	assert := assert.New(t)
	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-environment-dev").Return(nil).Once()
	stackManager.On("AwaitFinalStatus", "mu-loadbalancer-dev").Return(nil).Once()

	workflow := new(serviceWorkflow)
	err := workflow.serviceEnvironmentLoader("mu", "dev", stackManager)()

	assert.NotNil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
}

func TestServiceGetMaxPriority(t *testing.T) {
	assert := assert.New(t)

	elbRuleLister := new(mockedElbManager)
	elbRuleLister.On("ListRules", "foo").Return([]common.ElbRule{
		{Priority: stringRef("15")},
		{Priority: stringRef("5")},
		{Priority: stringRef("10")},
	})

	max := getMaxPriority(elbRuleLister, "foo")

	assert.Equal(15, max)

	elbRuleLister.AssertExpectations(t)
	elbRuleLister.AssertNumberOfCalls(t, "ListRules", 1)

}

func TestServiceEcsDeployer(t *testing.T) {
	assert := assert.New(t)

	stackManager := new(mockedStackManagerForService)
	stackManager.On("AwaitFinalStatus", "mu-service-foo-dev").Return(&common.Stack{Status: common.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-service-foo-dev").Return(nil)

	config := new(common.Config)
	config.Service.Name = "foo"

	params := make(map[string]string)

	workflow := new(serviceWorkflow)
	workflow.serviceName = "foo"
	outputs := make(map[string]string)
	outputs["provider"] = "ecs"
	workflow.envStack = &common.Stack{Name: "mu-environment-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	workflow.lbStack = &common.Stack{Name: "mu-loadbalancer-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	err := workflow.serviceEcsDeployer("mu", &config.Service, params, "dev", stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)

}

// mockKubernetesResourceManager mocks common/kubernetes.go:KubernetesResourceManager
type mockKubernetesResourceManager struct {
	mock.Mock
}

func (m *mockKubernetesResourceManager) UpsertResources(templateName string, templateData interface{}) error {
	args := m.Called(templateName)
	return args.Error(0)
}

func (m *mockKubernetesResourceManager) ListResources(apiVersion string, kind string, namespace string) (*unstructured.UnstructuredList, error) {
	args := m.Called(apiVersion)
	stack := args.Get(0)
	if stack == nil {
		return nil, args.Error(0)
	}
	return stack.(*unstructured.UnstructuredList), args.Error(0)
}

func (m *mockKubernetesResourceManager) DeleteResource(apiVersion string, kind string, namespace string, name string) error {
	args := m.Called(apiVersion)
	return args.Error(0)
}

// TestServiceEksDeployer tests that serviceWorkflow.serviceEksDeployer
// calls the kubernetesResourceManager.UpsertResources method once. It
// does not test the output of that call.
func TestServiceEksDeployer(t *testing.T) {
	assert := assert.New(t)

	// from workflows/service_common_test.go
	kubernetesResourceManager := new(mockKubernetesResourceManager)
	kubernetesResourceManager.On("UpsertResources", "kubernetes/deployment.yml").Return(nil)

	config := new(common.Config)
	config.Service.Name = "foo"

	params := make(map[string]string)

	workflow := new(serviceWorkflow)
	workflow.serviceName = "foo"
	outputs := make(map[string]string)
	outputs["provider"] = "eks"
	workflow.envStack = &common.Stack{Name: "mu-environment-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	workflow.lbStack = &common.Stack{Name: "mu-loadbalancer-dev", Status: common.StackStatusCreateComplete, Outputs: outputs}
	workflow.kubernetesResourceManager = kubernetesResourceManager
	err := workflow.serviceEksDeployer("mu", &config.Service, params, "dev")()
	assert.Nil(err)

	kubernetesResourceManager.AssertExpectations(t)
	kubernetesResourceManager.AssertNumberOfCalls(t, "UpsertResources", 1)
}

func stringRef(v string) *string {
	return &v
}

func TestServiceDeployer_serviceRolesetUpserter(t *testing.T) {
	assert := assert.New(t)
	rolesetManager := new(mockedRolesetManagerForService)

	rolesetManager.On("UpsertCommonRoleset").Return(nil)
	rolesetManager.On("GetCommonRoleset").Return(common.Roleset{"CloudFormationRoleArn": "bar"}, nil)
	rolesetManager.On("UpsertServiceRoleset", "env1", "svc20", "").Return(nil)
	rolesetManager.On("GetServiceRoleset").Return(common.Roleset{"EcsEventsRoleArn": "bar"}, nil)

	workflow := new(serviceWorkflow)
	workflow.serviceName = "svc20"
	err := workflow.serviceRolesetUpserter(rolesetManager, rolesetManager, "env1")()
	assert.Nil(err)
	assert.Equal("bar", workflow.cloudFormationRoleArn)

	rolesetManager.AssertExpectations(t)
	rolesetManager.AssertNumberOfCalls(t, "UpsertCommonRoleset", 1)
	rolesetManager.AssertNumberOfCalls(t, "GetCommonRoleset", 1)
	rolesetManager.AssertNumberOfCalls(t, "UpsertServiceRoleset", 1)

}
