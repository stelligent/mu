package common

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"io"
)

// StackWaiter for waiting on stack status to be final
type StackWaiter interface {
	AwaitFinalStatus(stackName string) string
}

// StackUpserter for applying changes to a stack
type StackUpserter interface {
	UpsertStack(stackName string, templateBodyReader io.Reader, stackParameters map[string]string) error
}

// StackManager composite of all stack capabilities
type StackManager interface {
	StackUpserter
	StackWaiter
}

type cloudformationStackManager struct {
	cfnAPI cloudformationiface.CloudFormationAPI
}

// TODO: support "dry-run" and write the template to a file
// fmt.Sprintf("%s/%s.yml", os.TempDir(), name),

// NewStackManager creates a new StackManager backed by cloudformation
func newStackManager(region string) (StackManager, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	cfn := cloudformation.New(sess, &aws.Config{Region: aws.String(region)})
	return &cloudformationStackManager{
		cfnAPI: cfn,
	}, nil
}

func buildStackParameters(stackParameters map[string]string) []*cloudformation.Parameter {
	parameters := make([]*cloudformation.Parameter, 0, len(stackParameters))
	for key, value := range stackParameters {
		parameters = append(parameters,
			&cloudformation.Parameter{
				ParameterKey:   aws.String(key),
				ParameterValue: aws.String(value),
			})
	}
	return parameters
}

// UpsertStack will create/update the cloudformation stack
func (cfnMgr *cloudformationStackManager) UpsertStack(stackName string, templateBodyReader io.Reader, stackParameters map[string]string) error {
	stackStatus := cfnMgr.AwaitFinalStatus(stackName)

	// load the template
	templateBodyBytes := new(bytes.Buffer)
	templateBodyBytes.ReadFrom(templateBodyReader)
	templateBody := aws.String(templateBodyBytes.String())

	cfnAPI := cfnMgr.cfnAPI
	if stackStatus == "" {
		fmt.Printf("creating stack: %s\n", stackName)
		params := &cloudformation.CreateStackInput{
			StackName: aws.String(stackName),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			Parameters:   buildStackParameters(stackParameters),
			TemplateBody: templateBody,
		}
		_, err := cfnAPI.CreateStack(params)
		if err != nil {
			return err
		}

		waitParams := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}
		cfnAPI.WaitUntilStackExists(waitParams)

	} else {
		fmt.Printf("updating stack: %s\n", stackName)
		params := &cloudformation.UpdateStackInput{
			StackName: aws.String(stackName),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			Parameters:   buildStackParameters(stackParameters),
			TemplateBody: templateBody,
		}

		_, err := cfnAPI.UpdateStack(params)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ValidationError" && awsErr.Message() == "No updates are to be performed." {
					fmt.Printf("No changes for stack: %s\n", stackName)
					return nil
				}
			}
			fmt.Println(err)
			return err
		}

	}
	return nil
}

// AwaitFinalStatus waits for the stack to arrive in a final status
//  returns: final status, or empty string if stack doesn't exist
func (cfnMgr *cloudformationStackManager) AwaitFinalStatus(stackName string) string {
	cfnAPI := cfnMgr.cfnAPI
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	resp, err := cfnAPI.DescribeStacks(params)

	if err == nil && resp != nil && len(resp.Stacks) == 1 {
		switch *resp.Stacks[0].StackStatus {
		case cloudformation.StackStatusReviewInProgress,
			cloudformation.StackStatusCreateInProgress,
			cloudformation.StackStatusRollbackInProgress:
			// wait for create
			cfnAPI.WaitUntilStackCreateComplete(params)
			resp, err = cfnAPI.DescribeStacks(params)
		case cloudformation.StackStatusDeleteInProgress:
			// wait for delete
			cfnAPI.WaitUntilStackDeleteComplete(params)
			resp, err = cfnAPI.DescribeStacks(params)
		case cloudformation.StackStatusUpdateInProgress,
			cloudformation.StackStatusUpdateRollbackInProgress,
			cloudformation.StackStatusUpdateCompleteCleanupInProgress,
			cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress:
			// wait for update
			cfnAPI.WaitUntilStackUpdateComplete(params)
			resp, err = cfnAPI.DescribeStacks(params)
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
