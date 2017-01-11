package common

import (
	"fmt"
	"os"
	"text/template"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Stack contains the data about a CloudFormation stack
type Stack struct {
	Name string
	TemplatePath string
}

// NewStack will create a new stack instance
func NewStack(name string) *Stack {
	return &Stack{
		Name: name,
		TemplatePath: fmt.Sprintf("%s/%s.yml",os.TempDir(), name),
	}
}
func newCloudFormation(region string) (*cloudformation.CloudFormation, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return cloudformation.New(sess, &aws.Config{Region: aws.String(region)}), nil
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

// UpsertStack will create/update the cloudformation stack
func (stack *Stack) UpsertStack(cfn *cloudformation.CloudFormation) (error) {
	stackStatus := stack.AwaitFinalStatus(cfn)
	if stackStatus == "" {
		fmt.Printf("creating stack: %s\n", stack.Name)
		params := &cloudformation.CreateStackInput{
			StackName: aws.String(stack.Name),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			TemplateBody: aws.String(stack.readTemplatePath()),
		}
		resp, err := cfn.CreateStack(params)
		if err != nil {
			return err
		}

		fmt.Println(resp)
	} else {
		fmt.Printf("updating stack: %s\n", stack.Name)
		params := &cloudformation.UpdateStackInput{
			StackName: aws.String(stack.Name),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			TemplateBody: aws.String(stack.readTemplatePath()),
		}

		resp, err := cfn.UpdateStack(params)
		if err != nil {
			return err
		}

		fmt.Println(resp)
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
func (stack *Stack) AwaitFinalStatus(cfn *cloudformation.CloudFormation) (string) {
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stack.Name),
	}
	resp, err := cfn.DescribeStacks(params)

	if err == nil && resp != nil && len(resp.Stacks) == 1 {
		switch *resp.Stacks[0].StackStatus {
			case cloudformation.StackStatusCreateFailed:
			case cloudformation.StackStatusCreateComplete:
			case cloudformation.StackStatusRollbackFailed:
			case cloudformation.StackStatusRollbackComplete:
			case cloudformation.StackStatusDeleteFailed:
			case cloudformation.StackStatusDeleteComplete:
			case cloudformation.StackStatusUpdateComplete:
			case cloudformation.StackStatusUpdateRollbackFailed:
			case cloudformation.StackStatusUpdateRollbackComplete:
				break;

			case cloudformation.StackStatusReviewInProgress:
			case cloudformation.StackStatusCreateInProgress:
			case cloudformation.StackStatusRollbackInProgress:
				// wait for create
				cfn.WaitUntilStackCreateComplete(params)
				resp, err = cfn.DescribeStacks(params)
				break;
			case cloudformation.StackStatusDeleteInProgress:
				// wait for delete
				cfn.WaitUntilStackDeleteComplete(params)
				resp, err = cfn.DescribeStacks(params)
				break;
			case cloudformation.StackStatusUpdateInProgress:
			case cloudformation.StackStatusUpdateRollbackInProgress:
			case cloudformation.StackStatusUpdateCompleteCleanupInProgress:
			case cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress:
				// wait for update
				cfn.WaitUntilStackUpdateComplete(params)
				resp, err = cfn.DescribeStacks(params)
				break;
		}
		return *resp.Stacks[0].StackStatus
	}

	return ""
}
