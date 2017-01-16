package common

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
	"fmt"
)

type mockedCloudFormation struct {
	mock.Mock
	cloudformationiface.CloudFormationAPI
}

func (m *mockedCloudFormation) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	args := m.Called()
	return args.Get(0).(*cloudformation.DescribeStacksOutput), args.Error(1)
}
func (m *mockedCloudFormation) DescribeStacksPages(input *cloudformation.DescribeStacksInput, cb func(*cloudformation.DescribeStacksOutput, bool) bool) error {
	args := m.Called(input, cb)
	return args.Error(0)
}

func (m *mockedCloudFormation) WaitUntilStackCreateComplete(*cloudformation.DescribeStacksInput) error {
	m.Called()
	return nil
}
func (m *mockedCloudFormation) WaitUntilStackUpdateComplete(*cloudformation.DescribeStacksInput) error {
	m.Called()
	return nil
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

	finalStatus := stackManager.AwaitFinalStatus("foo")

	assert.Equal(cloudformation.StackStatusCreateComplete, finalStatus)
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
	cfn.On("DescribeStacks").Return(
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		}, nil)

	cfn.On("WaitUntilStackCreateComplete").Return(nil)

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}

	finalStatus := stackManager.AwaitFinalStatus("foo")

	assert.Equal(cloudformation.StackStatusCreateComplete, finalStatus)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 2)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackCreateComplete", 1)
}

func TestStack_UpsertStack_Create(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacks").Return(&cloudformation.DescribeStacksOutput{}, errors.New("stack not found"))
	cfn.On("CreateStack").Return(&cloudformation.CreateStackOutput{}, nil)
	cfn.On("WaitUntilStackExists").Return(nil)

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}
	err := stackManager.UpsertStack("foo", strings.NewReader(""), nil, nil)

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
	cfn.AssertNumberOfCalls(t, "CreateStack", 1)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackExists", 1)
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

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}
	err := stackManager.UpsertStack("foo", strings.NewReader(""), nil, nil)

	assert.Nil(err)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacks", 1)
	cfn.AssertNumberOfCalls(t, "CreateStack", 0)
	cfn.AssertNumberOfCalls(t, "UpdateStack", 1)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackUpdateComplete", 0)
	cfn.AssertNumberOfCalls(t, "WaitUntilStackExists", 0)
}

func TestCloudformationStackManager_ListStacks(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.On("DescribeStacksPages", mock.AnythingOfType("*cloudformation.DescribeStacksInput"), mock.AnythingOfType("func(*cloudformation.DescribeStacksOutput, bool) bool")).
		Return(nil).
		Run(func (args mock.Arguments) {
		fmt.Println(args)
		cb := args.Get(1).(func(*cloudformation.DescribeStacksOutput, bool) bool)
		cb(&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				{
					StackName: aws.String("mu-cluster-dev"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Tags: []*cloudformation.Tag{
						{
							Key: aws.String("mu:type"),
							Value: aws.String("cluster"),
						},
					},
				},
				{
					StackName: aws.String("mu-vpc-dev"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Tags: []*cloudformation.Tag{
						{
							Key: aws.String("mu:type"),
							Value: aws.String("vpc"),
						},
					},
				},
				{
					StackName: aws.String("deleted-stack"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
					Tags: []*cloudformation.Tag{
						{
							Key: aws.String("mu:type"),
							Value: aws.String("cluster"),
						},
					},
				},
			},
		},true)
	})

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}
	stacks, err := stackManager.ListStacks(StackTypeCluster)

	assert.Nil(err)
	assert.NotNil(stacks)
	assert.Equal(1,len(stacks))
	assert.Equal("mu-cluster-dev", stacks[0].Name)
	assert.Equal("cluster", stacks[0].Tags["type"])
	assert.Equal(cloudformation.StackStatusCreateComplete, stacks[0].Status)
	cfn.AssertExpectations(t)
	cfn.AssertNumberOfCalls(t, "DescribeStacksPages", 1)
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
	assert.Equal(2, len(parameters))

	paramMap["p1"] = "value 1"
	paramMap["p2"] = "value 2"
	parameters = buildStackTags(paramMap)
	assert.Equal(4, len(parameters))
	assert.Contains(*parameters[0].Key, "mu:")
	assert.Contains(*parameters[1].Key, "mu:")
	assert.Contains(*parameters[2].Key, "mu:")
	assert.Contains(*parameters[3].Key, "mu:")
}
