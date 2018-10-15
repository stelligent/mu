package aws

import (
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedRolesetStackManager struct {
	mock.Mock
	common.StackManager
}

func (m *mockedRolesetStackManager) UpsertStack(stackName string, templateName string, templateData interface{}, parameters map[string]string, tags map[string]string, policy string, roleArn string) error {
	args := m.Called(stackName)
	return args.Error(0)
}

func (m *mockedRolesetStackManager) DeleteStack(stackName string) error {
	args := m.Called(stackName)
	return args.Error(0)
}

func (m *mockedRolesetStackManager) AwaitFinalStatus(stackName string) *common.Stack {
	args := m.Called(stackName)
	rtn := args.Get(0)
	if rtn == nil {
		return nil
	}
	return rtn.(*common.Stack)
}

func (m *mockedRolesetStackManager) GetStack(stackName string) (*common.Stack, error) {
	args := m.Called(stackName)
	return args.Get(0).(*common.Stack), args.Error(1)
}

func (m *mockedRolesetStackManager) SetTerminationProtection(stackName string, enabled bool) error {
	args := m.Called(stackName, enabled)
	return args.Error(0)
}

func TestIamRolesetManager_GetCommonRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)
	stackManagerMock.On("AwaitFinalStatus", "n1-iam-common").Return(&common.Stack{
		Outputs: map[string]string{
			"CloudFormationRoleArn": "foo",
		},
	}, nil)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "n1",
			},
		},
	}

	roleset, err := i.GetCommonRoleset()
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(1, len(roleset))
	assert.Equal("foo", roleset["CloudFormationRoleArn"])

	i.context.Config.Roles.CloudFormation = "bar"
	roleset, err = i.GetCommonRoleset()
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(1, len(roleset))
	assert.Equal("bar", roleset["CloudFormationRoleArn"])

}

func TestIamRolesetManager_GetEnvironmentRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)
	stackManagerMock.On("AwaitFinalStatus", "n2-iam-environment-env1").Return(&common.Stack{
		Outputs: map[string]string{
			"EC2InstanceProfileArn": "foo1",
		},
	}, nil)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "n2",
			},
		},
	}

	roleset, err := i.GetEnvironmentRoleset("env1")
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(1, len(roleset))
	assert.Equal("foo1", roleset["EC2InstanceProfileArn"])

	i.context.Config.Environments = []common.Environment{
		{
			Name: "env1",
		},
		{
			Name: "env2",
		},
	}
	i.context.Config.Environments[0].Roles.Instance = "bar1"
	i.context.Config.Environments[1].Roles.Instance = "bar2"

	roleset, err = i.GetEnvironmentRoleset("env1")
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(1, len(roleset))
	assert.Equal("bar1", roleset["EC2InstanceProfileArn"])

}

func TestIamRolesetManager_GetServiceRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-service-s1-env1").Return(&common.Stack{
		Outputs: map[string]string{
			"EcsServiceRoleArn": "foo3",
			"EcsTaskRoleArn":    "foo4",
		},
	}, nil)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
			},
		},
	}

	roleset, err := i.GetServiceRoleset("env1", "s1")
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(2, len(roleset))
	assert.Equal("foo3", roleset["EcsServiceRoleArn"])
	assert.Equal("foo4", roleset["EcsTaskRoleArn"])

	i.context.Config.Service.Roles.EcsService = "bar3"
	i.context.Config.Service.Roles.EcsTask = "bar4"

	roleset, err = i.GetServiceRoleset("env1", "s1")
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(2, len(roleset))
	assert.Equal("bar3", roleset["EcsServiceRoleArn"])
}

func TestIamRolesetManager_GetPipelineRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-pipeline-s1").Return(&common.Stack{
		Outputs: map[string]string{
			"CodePipelineRoleArn": "foo5",
			"MuAcptRoleArn":       "foo6",
		},
	}, nil)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
			},
		},
	}

	roleset, err := i.GetPipelineRoleset("s1")
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(2, len(roleset))
	assert.Equal("foo5", roleset["CodePipelineRoleArn"])
	assert.Equal("foo6", roleset["MuAcptRoleArn"])

	i.context.Config.Service.Pipeline.Roles.Pipeline = "bar5"
	i.context.Config.Service.Pipeline.Acceptance.Roles.Mu = "bar6"

	roleset, err = i.GetPipelineRoleset("s1")
	assert.Nil(err)
	assert.NotNil(roleset)
	stackManagerMock.AssertExpectations(t)
	assert.Equal(2, len(roleset))
	assert.Equal("bar5", roleset["CodePipelineRoleArn"])
	assert.Equal("bar6", roleset["MuAcptRoleArn"])
}

func TestIamRolesetManager_UpsertCommonRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace:  "mu",
				DisableIAM: true,
			},
		},
	}

	err := i.UpsertCommonRoleset()
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "UpsertStack", 0)

	stackManagerMock.On("UpsertStack", "mu-iam-common").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-common").Return(&common.Stack{Status: "CREATE_COMPLETE"})
	stackManagerMock.On("SetTerminationProtection", "mu-iam-common", true).Return(nil)
	i.context.Config.DisableIAM = false
	err = i.UpsertCommonRoleset()
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "UpsertStack", 1)
	stackManagerMock.AssertNumberOfCalls(t, "SetTerminationProtection", 1)
}

func TestIamRolesetManager_UpsertEnvironmentRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace:  "mu",
				DisableIAM: false,
			},
		},
	}

	err := i.UpsertEnvironmentRoleset("env1")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "UpsertStack", 0)

	i.context.Config.Environments = []common.Environment{
		{
			Name: "env1",
		},
		{
			Name: "env2",
		},
	}

	stackManagerMock.On("UpsertStack", "mu-iam-environment-env1").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-environment-env1").Return(&common.Stack{Status: "CREATE_COMPLETE"})
	err = i.UpsertEnvironmentRoleset("env1")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestIamRolesetManager_UpsertServiceRoleset_ManagedEnv(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
				Service: common.Service{
					Name: "sv1",
				},
			},
		},
	}

	i.context.Config.Environments = []common.Environment{
		{
			Name: "env1",
		},
		{
			Name: "env2",
		},
	}

	stackManagerMock.On("UpsertStack", "mu-iam-service-sv1-env1").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-service-sv1-env1").Return(&common.Stack{Status: "CREATE_COMPLETE"})

	err := i.UpsertServiceRoleset("env1", "sv1", "foo-bucket", "")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManagerMock.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestIamRolesetManager_UpsertServiceRoleset_SharedEnv(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
				Service: common.Service{
					Name: "sv1",
				},
			},
		},
	}
	stackManagerMock.On("AwaitFinalStatus", "mu-environment-env1").Return(&common.Stack{
		Status: "CREATE_COMPLETE",
		Tags: map[string]string{
			"provider": "ec2",
		},
	}, nil)
	stackManagerMock.On("UpsertStack", "mu-iam-service-sv1-env1").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-service-sv1-env1").Return(&common.Stack{Status: "CREATE_COMPLETE"})
	err := i.UpsertServiceRoleset("env1", "sv1", "foo-bucket", "")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	stackManagerMock.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestIamRolesetManager_UpsertPipelineRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
				Service: common.Service{
					Name: "sv1",
				},
			},
		},
	}
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-common").Return(&common.Stack{
		Outputs: map[string]string{
			"CloudFormationRoleArn": "foo",
		},
	}, nil)
	stackManagerMock.On("UpsertStack", "mu-iam-pipeline-sv1").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-pipeline-sv1").Return(&common.Stack{Status: "CREATE_COMPLETE"})
	err := i.UpsertPipelineRoleset("sv1", "test-bucket", "foo-bucket")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "AwaitFinalStatus", 2)
	stackManagerMock.AssertNumberOfCalls(t, "UpsertStack", 1)
}

func TestIamRolesetManager_DeleteCommonRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
			},
		},
	}
	stackManagerMock.On("DeleteStack", "mu-iam-common").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-common").Return(nil)
	stackManagerMock.On("SetTerminationProtection", "mu-iam-common", false).Return(nil)
	err := i.DeleteCommonRoleset()
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManagerMock.AssertNumberOfCalls(t, "DeleteStack", 1)
	stackManagerMock.AssertNumberOfCalls(t, "SetTerminationProtection", 1)
}

func TestIamRolesetManager_DeleteEnvironmentRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
			},
		},
	}
	stackManagerMock.On("DeleteStack", "mu-iam-environment-env10").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-environment-env10").Return(nil)
	err := i.DeleteEnvironmentRoleset("env10")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManagerMock.AssertNumberOfCalls(t, "DeleteStack", 1)
}

func TestIamRolesetManager_DeleteServiceRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
				Service: common.Service{
					Name: "sv1",
				},
			},
		},
	}
	stackManagerMock.On("DeleteStack", "mu-iam-service-sv1-env10").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-service-sv1-env10").Return(nil)
	err := i.DeleteServiceRoleset("env10", "sv1")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManagerMock.AssertNumberOfCalls(t, "DeleteStack", 1)
}

func TestIamRolesetManager_DeletePipelineRoleset(t *testing.T) {
	assert := assert.New(t)

	stackManagerMock := new(mockedRolesetStackManager)

	i := iamRolesetManager{
		context: &common.Context{
			StackManager: stackManagerMock,
			Config: common.Config{
				Namespace: "mu",
				Service: common.Service{
					Name: "sv1",
				},
			},
		},
	}
	stackManagerMock.On("DeleteStack", "mu-iam-pipeline-sv1").Return(nil)
	stackManagerMock.On("AwaitFinalStatus", "mu-iam-pipeline-sv1").Return(nil)
	err := i.DeletePipelineRoleset("sv1")
	assert.Nil(err)
	stackManagerMock.AssertExpectations(t)
	stackManagerMock.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManagerMock.AssertNumberOfCalls(t, "DeleteStack", 1)
}
