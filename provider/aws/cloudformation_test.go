package aws

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
	"time"
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
func (m *mockedCloudFormation) DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error) {
	args := m.Called()
	return args.Get(0).(*cloudformation.DescribeStackResourcesOutput), args.Error(1)
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
func (m *mockedCloudFormation) CreateStack(*cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	args := m.Called()
	return args.Get(0).(*cloudformation.CreateStackOutput), args.Error(1)
}
func (m *mockedCloudFormation) UpdateStack(*cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	args := m.Called()
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
	cfn.On("CreateStack").Return(&cloudformation.CreateStackOutput{}, nil)
	cfn.On("WaitUntilStackExists").Return(nil)

	extMgr := new(mockedExtensionsManager)
	extMgr.On("DecorateStackTemplate").Return()
	extMgr.On("DecorateStackParameters").Return()
	extMgr.On("DecorateStackTags").Return()

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
	}
	err := stackManager.UpsertStack("foo", "bucket.yml", nil, nil, nil, "")

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
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
	cfn.On("UpdateStack").Return(&cloudformation.UpdateStackOutput{}, nil)

	extMgr := new(mockedExtensionsManager)
	extMgr.On("DecorateStackTemplate").Return()
	extMgr.On("DecorateStackParameters").Return()
	extMgr.On("DecorateStackTags").Return()

	stackManager := cloudformationStackManager{
		cfnAPI:            cfn,
		extensionsManager: extMgr,
	}
	err := stackManager.UpsertStack("foo", "bucket.yml", nil, nil, nil, "")

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
	stacks, err := stackManager.ListStacks(common.StackTypeEnv)

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
	cfn.On("DescribeStackResources").Return(
		&cloudformation.DescribeStackResourcesOutput{
			StackResources: []*cloudformation.StackResource{
			},
		}, nil)

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
