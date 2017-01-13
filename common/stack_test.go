package common

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type mockedCloudFormation struct {
	cloudformationiface.CloudFormationAPI
	responses                    []*cloudformation.DescribeStacksOutput
	responseCount                int
	createCount                  int
	updateCount                  int
	waitUntilStackExists         int
	waitUntilStackCreateComplete int
	waitUntilStackUpdateComplete int
}

func (c *mockedCloudFormation) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if c.responseCount >= len(c.responses) {
		return nil, errors.New("stack not found")
	}

	resp := c.responses[c.responseCount]
	c.responseCount = c.responseCount + 1
	if resp == nil {
		return nil, errors.New("stack not found")
	}
	return resp, nil
}

func (c *mockedCloudFormation) WaitUntilStackCreateComplete(*cloudformation.DescribeStacksInput) error {
	c.waitUntilStackCreateComplete = c.waitUntilStackCreateComplete + 1
	return nil
}
func (c *mockedCloudFormation) WaitUntilStackUpdateComplete(*cloudformation.DescribeStacksInput) error {
	c.waitUntilStackUpdateComplete = c.waitUntilStackUpdateComplete + 1
	return nil
}
func (c *mockedCloudFormation) WaitUntilStackExists(*cloudformation.DescribeStacksInput) error {
	c.waitUntilStackExists = c.waitUntilStackExists + 1
	return nil
}
func (c *mockedCloudFormation) CreateStack(*cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	c.createCount = c.createCount + 1
	return nil, nil
}
func (c *mockedCloudFormation) UpdateStack(*cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	c.updateCount = c.updateCount + 1
	return nil, nil
}

func TestStack_AwaitFinalStatus_CreateComplete(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.responses = []*cloudformation.DescribeStacksOutput{
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		},
	}

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}

	finalStatus := stackManager.AwaitFinalStatus("foo")

	assert.Equal(cloudformation.StackStatusCreateComplete, finalStatus)
	assert.Equal(1, cfn.responseCount)
}

func TestStack_AwaitFinalStatus_CreateInProgress(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.responses = []*cloudformation.DescribeStacksOutput{
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String(cloudformation.StackStatusCreateInProgress),
				},
			},
		},
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		},
	}

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}

	finalStatus := stackManager.AwaitFinalStatus("foo")

	assert.Equal(cloudformation.StackStatusCreateComplete, finalStatus)
	assert.Equal(2, cfn.responseCount)
	assert.Equal(1, cfn.waitUntilStackCreateComplete)
}

func TestStack_UpsertStack_Create(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.responses = []*cloudformation.DescribeStacksOutput{
		nil,
	}

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}
	err := stackManager.UpsertStack("foo", strings.NewReader(""), nil, nil)

	assert.Nil(err)
	assert.Equal(1, cfn.waitUntilStackExists)
	assert.Equal(1, cfn.createCount)
	assert.Equal(1, cfn.responseCount)
}

func TestStack_UpsertStack_Update(t *testing.T) {
	assert := assert.New(t)

	cfn := new(mockedCloudFormation)
	cfn.responses = []*cloudformation.DescribeStacksOutput{
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		},
	}

	stackManager := cloudformationStackManager{
		cfnAPI: cfn,
	}
	err := stackManager.UpsertStack("foo", strings.NewReader(""), nil, nil)

	assert.Nil(err)
	assert.Equal(0, cfn.waitUntilStackExists)
	assert.Equal(0, cfn.createCount)
	assert.Equal(1, cfn.updateCount)
	assert.Equal(0, cfn.waitUntilStackUpdateComplete)
	assert.Equal(1, cfn.responseCount)
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
