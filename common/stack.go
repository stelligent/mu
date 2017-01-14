package common

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/op/go-logging"
	"io"
	"strings"
	"time"
)

var log = logging.MustGetLogger("stack")

// CreateStackName will create a name for a stack
func CreateStackName(stackType StackType, name string) string {
	return fmt.Sprintf("mu-%s-%s", stackType, name)
}

// StackWaiter for waiting on stack status to be final
type StackWaiter interface {
	AwaitFinalStatus(stackName string) string
}

// StackUpserter for applying changes to a stack
type StackUpserter interface {
	UpsertStack(stackName string, templateBodyReader io.Reader, parameters map[string]string, tags map[string]string) error
}

// StackLister for listing stacks
type StackLister interface {
	ListStacks(stackType StackType) ([]*Stack, error)
}

// StackManager composite of all stack capabilities
type StackManager interface {
	StackUpserter
	StackWaiter
	StackLister
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
	log.Debugf("Connecting to CloudFormation service in region:%s", region)
	cfn := cloudformation.New(sess, &aws.Config{Region: aws.String(region)})
	return &cloudformationStackManager{
		cfnAPI: cfn,
	}, nil
}

func buildStackParameters(parameters map[string]string) []*cloudformation.Parameter {
	stackParameters := make([]*cloudformation.Parameter, 0, len(parameters))
	for key, value := range parameters {
		stackParameters = append(stackParameters,
			&cloudformation.Parameter{
				ParameterKey:   aws.String(key),
				ParameterValue: aws.String(value),
			})
	}
	return stackParameters
}

func buildStackTags(tags map[string]string) []*cloudformation.Tag {
	stackTags := make([]*cloudformation.Tag, 0, len(tags)+2)

	stackTags = append(stackTags,
		&cloudformation.Tag{
			Key:   aws.String("mu:version"),
			Value: aws.String(GetVersion()),
		},
		&cloudformation.Tag{
			Key:   aws.String("mu:lastupdate"),
			Value: aws.String(fmt.Sprintf("%v", time.Now().Unix())),
		})

	for key, value := range tags {
		stackTags = append(stackTags,
			&cloudformation.Tag{
				Key:   aws.String(fmt.Sprintf("mu:%s", key)),
				Value: aws.String(value),
			})
	}
	return stackTags
}

// UpsertStack will create/update the cloudformation stack
func (cfnMgr *cloudformationStackManager) UpsertStack(stackName string, templateBodyReader io.Reader, parameters map[string]string, tags map[string]string) error {
	stackStatus := cfnMgr.AwaitFinalStatus(stackName)

	// load the template
	templateBodyBytes := new(bytes.Buffer)
	templateBodyBytes.ReadFrom(templateBodyReader)
	templateBody := aws.String(templateBodyBytes.String())

	// stack parameters
	stackParameters := buildStackParameters(parameters)

	// stack tags
	stackTags := buildStackTags(tags)

	cfnAPI := cfnMgr.cfnAPI
	if stackStatus == "" {

		log.Debugf("  Creating stack named '%s'", stackName)
		log.Debugf("  Stack parameters:\n\t%s", stackParameters)
		log.Debugf("  Stack tags:\n\t%s", stackTags)
		params := &cloudformation.CreateStackInput{
			StackName: aws.String(stackName),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			Parameters:   stackParameters,
			TemplateBody: templateBody,
			Tags:         stackTags,
		}
		_, err := cfnAPI.CreateStack(params)
		log.Debug("  Create stack complete err=%s", err)
		if err != nil {
			return err
		}

		waitParams := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}
		log.Debug("  Waiting for stack to exist...")
		cfnAPI.WaitUntilStackExists(waitParams)
		log.Debug("  Stack exists.")

	} else {
		log.Debugf("  Updating stack named '%s'", stackName)
		log.Debugf("  Prior state: %s", stackStatus)
		log.Debugf("  Stack parameters:\n\t%s", stackParameters)
		log.Debugf("  Stack tags:\n\t%s", stackTags)
		params := &cloudformation.UpdateStackInput{
			StackName: aws.String(stackName),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			Parameters:   stackParameters,
			TemplateBody: templateBody,
			Tags:         stackTags,
		}

		_, err := cfnAPI.UpdateStack(params)
		log.Debug("  Update stack complete err=%s", err)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ValidationError" && awsErr.Message() == "No updates are to be performed." {
					log.Noticef("  No changes for stack '%s'", stackName)
					return nil
				}
			}
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
			log.Debugf("  Waiting for stack:%s to complete...current status=%s", stackName, *resp.Stacks[0].StackStatus)
			cfnAPI.WaitUntilStackCreateComplete(params)
			resp, err = cfnAPI.DescribeStacks(params)
		case cloudformation.StackStatusDeleteInProgress:
			// wait for delete
			log.Debugf("  Waiting for stack:%s to delete...current status=%s", stackName, *resp.Stacks[0].StackStatus)
			cfnAPI.WaitUntilStackDeleteComplete(params)
			resp, err = cfnAPI.DescribeStacks(params)
		case cloudformation.StackStatusUpdateInProgress,
			cloudformation.StackStatusUpdateRollbackInProgress,
			cloudformation.StackStatusUpdateCompleteCleanupInProgress,
			cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress:
			// wait for update
			log.Debugf("  Waiting for stack:%s to update...current status=%s", stackName, *resp.Stacks[0].StackStatus)
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
		log.Debugf("  Returning final status for stack:%s ... status=%s", stackName, *resp.Stacks[0].StackStatus)
		return *resp.Stacks[0].StackStatus
	}

	log.Debugf("  Stack doesn't exist ... stack=%s", stackName)
	return ""
}

// ListStacks will find mu stacks
func (cfnMgr *cloudformationStackManager) ListStacks(stackType StackType) ([]*Stack, error) {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.ListStacksInput{
		StackStatusFilter: []*string{
			aws.String(cloudformation.StackStatusReviewInProgress),
			aws.String(cloudformation.StackStatusCreateInProgress),
			aws.String(cloudformation.StackStatusRollbackInProgress),
			aws.String(cloudformation.StackStatusDeleteInProgress),
			aws.String(cloudformation.StackStatusUpdateInProgress),
			aws.String(cloudformation.StackStatusUpdateRollbackInProgress),
			aws.String(cloudformation.StackStatusUpdateCompleteCleanupInProgress),
			aws.String(cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress),
			aws.String(cloudformation.StackStatusCreateFailed),
			aws.String(cloudformation.StackStatusCreateComplete),
			aws.String(cloudformation.StackStatusRollbackFailed),
			aws.String(cloudformation.StackStatusRollbackComplete),
			aws.String(cloudformation.StackStatusDeleteFailed),
			aws.String(cloudformation.StackStatusUpdateComplete),
			aws.String(cloudformation.StackStatusUpdateRollbackFailed),
			aws.String(cloudformation.StackStatusUpdateRollbackComplete),
		},
	}

	var stacks []*Stack

	stackNamePrefix := "mu-"
	if stackType != "" {
		stackNamePrefix = fmt.Sprintf("%s%s-", stackNamePrefix, stackType)
	}
	log.Debugf("Searching for stacks with prefix '%s'", stackNamePrefix)

	err := cfnAPI.ListStacksPages(params,
		func(page *cloudformation.ListStacksOutput, lastPage bool) bool {
			for _, stackSummary := range page.StackSummaries {
				if !strings.HasPrefix(aws.StringValue(stackSummary.StackName), stackNamePrefix) {
					continue
				}
				if strings.HasPrefix(aws.StringValue(stackSummary.StackStatus), "DELETE_") {
					continue
				}
				stack := new(Stack)
				stack.ID = aws.StringValue(stackSummary.StackId)
				stack.Name = aws.StringValue(stackSummary.StackName)
				stack.Status = aws.StringValue(stackSummary.StackStatus)
				stack.StatusReason = aws.StringValue(stackSummary.StackStatusReason)

				stacks = append(stacks, stack)
			}
			return true
		})

	if err != nil {
		return nil, err
	}
	return stacks, nil
}
