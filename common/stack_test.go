package common

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/aws"
	"errors"
)

type mockedCloudFormation struct {
	cloudformationiface.CloudFormationAPI
	responses []*cloudformation.DescribeStacksOutput
	responseIndex int
	createIndex int
	updateIndex int
	waitUntilStackExists int
	waitUntilStackCreateComplete int
	waitUntilStackUpdateComplete int
}

func (c *mockedCloudFormation) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	if(c.responseIndex >= len(c.responses)) {
		return nil, errors.New("stack not found")
	}

	resp := c.responses[c.responseIndex]
	c.responseIndex = c.responseIndex + 1
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
	c.createIndex = c.createIndex + 1
	return nil, nil
}
func (c *mockedCloudFormation) UpdateStack(*cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	c.updateIndex = c.updateIndex + 1
	return nil, nil
}

func TestNewStack(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack("foo")

	assert.NotNil(stack)
	assert.Equal("foo",stack.Name)
}

func TestStack_WriteTemplate(t *testing.T) {
	assert := assert.New(t)

	stack := NewStack("foo")
	env := &Environment{}
	err := stack.WriteTemplate("environment-template.yml", env)
	assert.Nil(err)

	template := stack.readTemplatePath()
	assert.NotNil(template)
}

func TestStack_AwaitFinalStatus_CreateComplete(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack("foo")


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

	finalStatus := stack.AwaitFinalStatus(cfn)

	assert.Equal(cloudformation.StackStatusCreateComplete, finalStatus)
	assert.Equal(1, cfn.responseIndex)
}

func TestStack_AwaitFinalStatus_CreateInProgress(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack("foo")


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

	finalStatus := stack.AwaitFinalStatus(cfn)

	assert.Equal(cloudformation.StackStatusCreateComplete, finalStatus)
	assert.Equal(2, cfn.responseIndex)
	assert.Equal(1, cfn.waitUntilStackCreateComplete)
}

func TestStack_UpsertStack_Create(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack("foo")

	cfn := new(mockedCloudFormation)
	cfn.responses = []*cloudformation.DescribeStacksOutput{
		nil,
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

	err := stack.UpsertStack(cfn)

	assert.Nil(err)
	assert.Equal(1, cfn.waitUntilStackExists)
	assert.Equal(1, cfn.createIndex)
	assert.Equal(1, cfn.waitUntilStackCreateComplete)
	assert.Equal(3, cfn.responseIndex)
}

func TestStack_UpsertStack_Update(t *testing.T) {
	assert := assert.New(t)
	stack := NewStack("foo")

	cfn := new(mockedCloudFormation)
	cfn.responses = []*cloudformation.DescribeStacksOutput{
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		},
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String(cloudformation.StackStatusUpdateInProgress),
				},
			},
		},
		&cloudformation.DescribeStacksOutput{
			Stacks: []*cloudformation.Stack{
				&cloudformation.Stack{
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
				},
			},
		},
	}

	err := stack.UpsertStack(cfn)

	assert.Nil(err)
	assert.Equal(0, cfn.waitUntilStackExists)
	assert.Equal(0, cfn.createIndex)
	assert.Equal(1, cfn.updateIndex)
	assert.Equal(1, cfn.waitUntilStackUpdateComplete)
	assert.Equal(3, cfn.responseIndex)
}
