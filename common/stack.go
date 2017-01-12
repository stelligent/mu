package common

import (
	"fmt"
	"os"
	"text/template"
	"github.com/aws/aws-sdk-go/aws"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

// Stack contains the data about a CloudFormation stack
type Stack struct {
	Name string
	Parameters map[string]string
	TemplatePath string
}

// NewStack will create a new stack instance
func NewStack(name string) *Stack {
	return &Stack{
		Name: name,
		Parameters: make(map[string]string),
		TemplatePath: fmt.Sprintf("%s/%s.yml",os.TempDir(), name),
	}
}
func newCloudFormation(region string) (cloudformationiface.CloudFormationAPI, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return cloudformation.New(sess, &aws.Config{Region: aws.String(region)}), nil
}

// WithParameter apply a parameter to the stack
func (stack *Stack) WithParameter(key string, value string) (*Stack) {
	stack.Parameters[key] = value
	return stack
}

// WriteTemplate will create a temp file with the template for a CFN stack
//go:generate go-bindata -pkg $GOPACKAGE -o assets.go assets/
func (stack *Stack) WriteTemplate(assetName string, data interface{}) (error) {
	asset, err := Asset(fmt.Sprintf("assets/%s",assetName))
	if err != nil {
		return err
	}

	tmpl, err := template.New(assetName).Parse(string(asset[:]))
	if err != nil {
		return err
	}

	templateOut, err := os.Create(stack.TemplatePath)
	defer templateOut.Close()
	if err != nil {
		return err
	}

	err = tmpl.Execute(templateOut, data)
	if err != nil {
		return err
	}

	templateOut.Sync()
	return nil
}

func (stack *Stack) buildParameters() ([]*cloudformation.Parameter) {
	parameters := make([]*cloudformation.Parameter, 0, len(stack.Parameters))
	for key, value := range stack.Parameters {
		parameters = append(parameters,
			&cloudformation.Parameter {
				ParameterKey: aws.String(key),
				ParameterValue: aws.String(value),
			})
	}
	return parameters
}

// UpsertStack will create/update the cloudformation stack
func (stack *Stack) UpsertStack(cfn cloudformationiface.CloudFormationAPI) (error) {
	stackStatus := stack.AwaitFinalStatus(cfn)
	if stackStatus == "" {
		fmt.Printf("creating stack: %s\n", stack.Name)
		params := &cloudformation.CreateStackInput{
			StackName: aws.String(stack.Name),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			Parameters: stack.buildParameters(),
			TemplateBody: aws.String(stack.readTemplatePath()),
		}
		_, err := cfn.CreateStack(params)
		if err != nil {
			return err
		}

		waitParams := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stack.Name),
		}
		cfn.WaitUntilStackExists(waitParams)

	} else {
		fmt.Printf("updating stack: %s\n", stack.Name)
		params := &cloudformation.UpdateStackInput{
			StackName: aws.String(stack.Name),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			Parameters: stack.buildParameters(),
			TemplateBody: aws.String(stack.readTemplatePath()),
		}

		_, err := cfn.UpdateStack(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ValidationError" && awsErr.Message() == "No updates are to be performed." {
					fmt.Printf("No changes for stack: %s\n", stack.Name)
					return nil
				}
			}
			fmt.Println(err)
			return err
		}

	}
	stack.AwaitFinalStatus(cfn)
	return nil
}

func (stack *Stack) readTemplatePath() (string) {
	templateBytes, err := ioutil.ReadFile(stack.TemplatePath)
	if err != nil {
		return ""
	}
	return string(templateBytes)
}

// AwaitFinalStatus waits for the stack to arrive in a final status
func (stack *Stack) AwaitFinalStatus(cfn cloudformationiface.CloudFormationAPI) (string) {
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stack.Name),
	}
	resp, err := cfn.DescribeStacks(params)

	if err == nil && resp != nil && len(resp.Stacks) == 1 {
		switch *resp.Stacks[0].StackStatus {
		case cloudformation.StackStatusReviewInProgress,
		cloudformation.StackStatusCreateInProgress,
		cloudformation.StackStatusRollbackInProgress:
			// wait for create
			cfn.WaitUntilStackCreateComplete(params)
			resp, err = cfn.DescribeStacks(params)
		case cloudformation.StackStatusDeleteInProgress:
			// wait for delete
			cfn.WaitUntilStackDeleteComplete(params)
			resp, err = cfn.DescribeStacks(params)
		case cloudformation.StackStatusUpdateInProgress,
		cloudformation.StackStatusUpdateRollbackInProgress,
		cloudformation.StackStatusUpdateCompleteCleanupInProgress,
		cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress:
			// wait for update
			cfn.WaitUntilStackUpdateComplete(params)
			resp, err = cfn.DescribeStacks(params)
		case cloudformation.StackStatusCreateFailed,
		cloudformation.StackStatusCreateComplete,
		cloudformation.StackStatusRollbackFailed,
		cloudformation.StackStatusRollbackComplete,
		cloudformation.StackStatusDeleteFailed,
		cloudformation.StackStatusDeleteComplete,
		cloudformation.StackStatusUpdateComplete,
		cloudformation.StackStatusUpdateRollbackFailed,
		cloudformation.StackStatusUpdateRollbackComplete:
			// no op

		}
		return *resp.Stacks[0].StackStatus
	}

	return ""
}
