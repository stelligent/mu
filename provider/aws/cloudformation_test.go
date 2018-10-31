package aws

import (
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedExtensionsManager struct {
	mock.Mock
	common.ExtensionsManager
}

func (m *mockedExtensionsManager) DecorateStackTemplate(assetName string, stackName string, templateBody io.Reader) (io.Reader, error) {
	m.Called()
	return templateBody, nil
}
func (m *mockedExtensionsManager) DecorateStackParameters(stackName string, stackParameters map[string]string) (map[string]string, error) {
	m.Called()
	return stackParameters, nil
}
func (m *mockedExtensionsManager) DecorateStackTags(stackName string, stackTags map[string]string) (map[string]string, error) {
	m.Called()
	return stackTags, nil
}

type mockedCloudFormation struct {
	mock.Mock
	cloudformationiface.CloudFormationAPI
}

func (m *mockedCloudFormation) DeleteStack(input *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	args := m.Called()
	return args.Get(0).(*cloudformation.DeleteStackOutput), args.Error(1)
}
func (m *mockedCloudFormation) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	args := m.Called()
	return args.Get(0).(*cloudformation.DescribeStacksOutput), args.Error(1)
}
func (m *mockedCloudFormation) DescribeStackEvents(input *cloudformation.DescribeStackEventsInput) (*cloudformation.DescribeStackEventsOutput, error) {
	args := m.Called()
	return args.Get(0).(*cloudformation.DescribeStackEventsOutput), args.Error(1)
}
func (m *mockedCloudFormation) DescribeStacksPages(input *cloudformation.DescribeStacksInput, cb func(*cloudformation.DescribeStacksOutput, bool) bool) error {
	args := m.Called(input, cb)
	return args.Error(0)
}
func (m *mockedCloudFormation) WaitUntilStackExists(*cloudformation.DescribeStacksInput) error {
	m.Called()
	return nil
}
func (m *mockedCloudFormation) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudformation.CreateStackOutput), args.Error(1)
}
func (m *mockedCloudFormation) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*cloudformation.UpdateStackOutput), args.Error(1)
}

func TestStack_AwaitFinalStatus_CreateComplete(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacks").Return(
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}

	stack := stackManager.AwaitFinalStatus("foo")

	assert.Equal(cloudformation.StackStatusCreateComplete, stack.Status)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
}

func TestStack_AwaitFinalStatus_CreateInProgress(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacks").Return(
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: aws.String(cloudformation.StackStatusCreateInProgress),
				},
			},
		}, nil).Once()
	cfn.On("DescribeStackEvents").Return(
		&cloudformation.DescribeStackEventsOutput{
			StackEvents: []*cloudformation.StackEvent{},
		}, nil)

	cfn.On("DescribeStacks").Return(
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}

	stack := stackManager.AwaitFinalStatus("foo")

	assert.Equal(cloudformation.StackStatusCreateComplete, stack.Status)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 2)
	cfn.AssertNumberOfCalls(t, "DescribeStackEvents", 1)
}

func TestStack_UpsertStack_Create(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacks").Return(&cloudformation.DescribeStacksOutput{}, errors.New("stack not found"))
	cfn.On("DescribeStacksPages", mock.AnythingOfType("*cloudformation.DescribeStacksInput"), mock.AnythingOfType("func(*cloudformation.DescribeStacksOutput, bool) bool")).
		Return(nil)
	cfn.On("CreateStack", mock.MatchedBy(
		func(params *cloudformation.CreateStackInput) bool {
			return true
		},
	)).Return(&cloudformation.CreateStackOutput{}, nil)
	cfn.On("WaitUntilStackExists").Return(nil)

	extMgr := new(mockedExtensionsManager)
	extMgr.On("DecorateStackTemplate").Return()
	extMgr.On("DecorateStackParameters").Return()
	extMgr.On("DecorateStackTags").Return()

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
	}
	err := stackManager.UpsertStack("foo", "cloudformation/bucket.yml", nil, nil, nil, "", "")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 2)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackExists", 1)
	cfn.AssertNumberOfCalls(t, "CreateStack", 1)
}

func TestStack_UpsertStack_Update(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacks").Return(
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil)
	cfn.On("UpdateStack", mock.MatchedBy(func(params *cloudformation.UpdateStackInput) bool {
		return true
	})).Return(&cloudformation.UpdateStackOutput{}, nil)

	extMgr := new(mockedExtensionsManager)
	extMgr.On("DecorateStackTemplate").Return()
	extMgr.On("DecorateStackParameters").Return()
	extMgr.On("DecorateStackTags").Return()

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
	}
	err := stackManager.UpsertStack("foo", "cloudformation/bucket.yml", nil, nil, nil, "", "")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
	cfn.AssertNumberOfCalls(t, "CreateStack", 0)
	cfn.AssertNumberOfCalls(t, "UpdateStack", 1)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackExists", 0)
}

func TestCloudformationStackManager_ListStacks(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacksPages", mock.AnythingOfType("*cloudformation.DescribeStacksInput"), mock.AnythingOfType("func(*cloudformation.DescribeStacksOutput, bool) bool")).
		Return(nil).
		Run(func(args mock.Arguments) {
			cb := args.Get(1).(func(*cloudformation.DescribeStacksOutput, bool) bool)
			cb(&cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					{
						StackName:   aws.String("mu-environment-dev"),
						StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
						Tags: []*cloudformation.Tag{
							{
								Key:   aws.String("mu:type"),
								Value: aws.String("environment"),
							},
						},
					},
					{
						StackName:   aws.String("mu-vpc-dev"),
						StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
						Tags: []*cloudformation.Tag{
							{
								Key:   aws.String("mu:type"),
								Value: aws.String("vpc"),
							},
						},
					},
					{
						StackName:   aws.String("deleted-stack"),
						StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
						Tags: []*cloudformation.Tag{
							{
								Key:   aws.String("mu:type"),
								Value: aws.String("environment"),
							},
						},
					},
				},
			}, true)
		})

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}
	stacks, err := stackManager.ListStacks(common.StackTypeEnv, "mu")

	assert.Nil(err)
	assert.NotNil(stacks)
	assert.Equal(1, len(stacks))
	assert.Equal("mu-environment-dev", stacks[0].Name)
	assert.Equal("environment", stacks[0].Tags["type"])
	assert.Equal(cloudformation.StackStatusCreateComplete, stacks[0].Status)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacksPages", 1)
}

func TestStack_GetStack(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacks").Return(
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}

	stack, err := stackManager.GetStack("foo")

	assert.Nil(err)
	assert.Equal(cloudformation.StackStatusCreateComplete, stack.Status)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
}
func TestStack_DeleteStack(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DeleteStack").Return(&cloudformation.DeleteStackOutput{}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}

	err := stackManager.DeleteStack("foo")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DeleteStack", 1)
}

func TestBuildStack(t *testing.T) {
	assert := assert.New(t)

	stackDetails := cloudformation.Stack{
		StackName:       aws.String("mu-environment-dev"),
		StackStatus:     aws.String(cloudformation.StackStatusCreateComplete),
		LastUpdatedTime: aws.Time(time.Now()),
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("mu:type"),
				Value: aws.String("environment"),
			},
		},
	}

	stack := buildStack(&stackDetails)

	assert.NotNil(stack)
	assert.Equal("mu-environment-dev", stack.Name)
	assert.Equal("environment", stack.Tags["type"])
	assert.Equal(cloudformation.StackStatusCreateComplete, stack.Status)
	assert.Equal(aws.TimeValue(stackDetails.LastUpdatedTime), stack.LastUpdateTime)
}

func TestBuildStack_NoUpdateTime(t *testing.T) {
	assert := assert.New(t)

	stackDetails := cloudformation.Stack{
		StackName:    aws.String("mu-environment-dev"),
		CreationTime: aws.Time(time.Now()),
		Tags:         []*cloudformation.Tag{},
	}

	stack := buildStack(&stackDetails)

	assert.NotNil(stack)
	assert.Equal(aws.TimeValue(stackDetails.CreationTime), stack.LastUpdateTime)
}

func TestBuildParameters(t *testing.T) {
	assert := assert.New(t)

	paramMap := make(map[string]string)

	parameters := buildStackParameters(paramMap)
	assert.Equal(0, len(parameters))

	paramMap["p1"] = "value 1"
	paramMap["p2"] = "value 2"
	parameters = buildStackParameters(paramMap)
	assert.Equal(2, len(parameters))
	assert.Contains(*parameters[0].ParameterKey, "p")
	assert.Contains(*parameters[0].ParameterValue, "value")
	assert.Contains(*parameters[1].ParameterKey, "p")
	assert.Contains(*parameters[1].ParameterValue, "value")
}

func TestTagParameters(t *testing.T) {
	assert := assert.New(t)

	paramMap := make(map[string]string)

	parameters := buildStackTags(paramMap)
	assert.Equal(1, len(parameters))

	paramMap["mu:p1"] = "value 1"
	paramMap["mu:p2"] = "value 2"
	parameters = buildStackTags(paramMap)
	assert.Equal(3, len(parameters))
	assert.Contains(*parameters[0].Key, "mu:")
	assert.Contains(*parameters[1].Key, "mu:")
	assert.Contains(*parameters[2].Key, "mu:")
}

func TestNewStackManager(t *testing.T) {
	assert := assert.New(t)

	extMagr := new(mockedExtensionsManager)

	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}
	sess, _ := session.NewSessionWithOptions(sessOptions)

	stackMgr, _ := newStackManager(sess, extMagr, "test", true, true)
	assert.NotNil(stackMgr)
}

func TestGetPolicy(t *testing.T) {
	assert := assert.New(t)

	policy, err := getPolicy("", true)
	var statements Statements
	json.Unmarshal([]byte(policy), &statements)
	assert.Nil(err)
	assert.Len(statements.Statements, 1)
	assert.Equal(statements.Statements[0].Effect, "Allow")
	assert.Equal(statements.Statements[0].Action, "Update:*")
	assert.Equal(statements.Statements[0].Resource, "*")
}

type Statement struct {
	Effect    string
	Action    string
	Principal string
	Resource  string
}
type Statements struct {
	Statements []Statement `json:"Statement"`
}

func TestStack_UpsertStack_CreatePolicy(t *testing.T) {
	assert := assert.New(t)

	cfn := mockBasicCfnAPI(true)
	extMgr := mockNilExtensionManager()

	templateName := "cloudformation/bucket.yml"
	var templateData interface{}
	stackName := "foo"

	policy, _ := templates.GetAsset(common.TemplatePolicyDefault)

	cfn.On("DescribeStacksPages", mock.AnythingOfType("*cloudformation.DescribeStacksInput"), mock.AnythingOfType("func(*cloudformation.DescribeStacksOutput, bool) bool")).
		Return(nil)
	cfn.On("CreateStack", mock.MatchedBy(
		func(params *cloudformation.CreateStackInput) bool {
			return *params.StackPolicyBody == policy
		},
	)).Return(&cloudformation.CreateStackOutput{
		StackId: aws.String("1"),
	}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
	}
	err := stackManager.UpsertStack(stackName, templateName, templateData, nil, nil, policy, "")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "CreateStack", 1)
}

func TestStack_UpsertStack_CreatePolicyAllowDataLoss(t *testing.T) {
	assert := assert.New(t)

	cfn := mockBasicCfnAPI(true)
	extMgr := mockNilExtensionManager()

	templateName := "cloudformation/bucket.yml"
	var templateData interface{}
	stackName := "foo"

	policy, _ := templates.GetAsset(common.TemplatePolicyDefault)

	cfn.On("DescribeStacksPages", mock.AnythingOfType("*cloudformation.DescribeStacksInput"), mock.AnythingOfType("func(*cloudformation.DescribeStacksOutput, bool) bool")).
		Return(nil)
	cfn.On("CreateStack", mock.MatchedBy(
		func(params *cloudformation.CreateStackInput) bool {
			return *params.StackPolicyBody == policy
		},
	)).Return(&cloudformation.CreateStackOutput{
		StackId: aws.String("1"),
	}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
		allowDataLoss:     true,
	}
	err := stackManager.UpsertStack(stackName, templateName, templateData, nil, nil, policy, "")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "CreateStack", 1)
}

func TestStack_UpsertStack_UpdatePolicy(t *testing.T) {
	assert := assert.New(t)

	cfn := mockBasicCfnAPI(false)
	extMgr := mockNilExtensionManager()

	templateName := "cloudformation/bucket.yml"
	var templateData interface{}
	stackName := "foo"

	policy, _ := templates.GetAsset(common.TemplatePolicyDefault)

	cfn.On("UpdateStack", mock.MatchedBy(
		func(params *cloudformation.UpdateStackInput) bool {
			return *params.StackPolicyDuringUpdateBody == policy
		},
	)).Return(&cloudformation.UpdateStackOutput{}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
	}
	err := stackManager.UpsertStack(stackName, templateName, templateData, nil, nil, policy, "")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
	cfn.AssertNumberOfCalls(t, "CreateStack", 0)
	cfn.AssertNumberOfCalls(t, "UpdateStack", 1)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackExists", 0)

}
func TestStack_UpsertStack_UpdatePolicyAllowDataLoss(t *testing.T) {
	assert := assert.New(t)

	cfn := mockBasicCfnAPI(false)
	extMgr := mockNilExtensionManager()

	templateName := "cloudformation/bucket.yml"
	var templateData interface{}
	stackName := "foo"

	allowDataLossPolicy, _ := templates.GetAsset(common.TemplatePolicyAllowAll)

	cfn.On("UpdateStack", mock.MatchedBy(
		func(params *cloudformation.UpdateStackInput) bool {
			return *params.StackPolicyDuringUpdateBody == allowDataLossPolicy
		},
	)).Return(&cloudformation.UpdateStackOutput{}, nil)

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
		allowDataLoss:     true,
	}

	policy, _ := templates.GetAsset(common.TemplatePolicyDefault)

	err := stackManager.UpsertStack(stackName, templateName, templateData, nil, nil, policy, "")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
	cfn.AssertNumberOfCalls(t, "CreateStack", 0)
	cfn.AssertNumberOfCalls(t, "UpdateStack", 1)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackExists", 0)

}

func mockBasicCfnAPI(create bool) *mockedCloudFormation {
	cfn := new(mockedCloudFormation)
	if create {
		cfn.On("DescribeStacks").Return(&cloudformation.DescribeStacksOutput{}, errors.New("stack not found"))
		cfn.On("WaitUntilStackExists").Return(nil)
	} else {
		cfn.On("DescribeStacks").Return(
			&cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					{
						StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					},
				},
			}, nil)
	}
	return cfn
}

func mockNilExtensionManager() *mockedExtensionsManager {
	extMgr := new(mockedExtensionsManager)
	extMgr.On("DecorateStackTemplate").Return()
	extMgr.On("DecorateStackParameters").Return()
	extMgr.On("DecorateStackTags").Return()
	return extMgr
}

func mockTemplateBody(extMgr *mockedExtensionsManager, stackName string, templateName string, templateData interface{}) *string {

	templateBody, _ := templates.GetAsset(templateName, templates.ExecuteTemplate(templateData),
		templates.DecorateTemplate(extMgr, stackName))

	return aws.String(templateBody)
}
