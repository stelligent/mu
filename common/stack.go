package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/briandowns/spinner"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// CreateStackName will create a name for a stack
func CreateStackName(stackType StackType, names ...string) string {
	return fmt.Sprintf("mu-%s-%s", stackType, strings.Join(names, "-"))
}

// GetStackOverrides will get the overrides from the config
func GetStackOverrides(stackName string) interface{} {
	if stackOverrides == nil {
		return nil
	}

	return stackOverrides[stackName]
}

var stackOverrides map[string]interface{}

func registerStackOverrides(overrides map[string]interface{}) {
	stackOverrides = overrides
}

// StackWaiter for waiting on stack status to be final
type StackWaiter interface {
	AwaitFinalStatus(stackName string) *Stack
}

// StackUpserter for applying changes to a stack
type StackUpserter interface {
	UpsertStack(stackName string, templateBodyReader io.Reader, parameters map[string]string, tags map[string]string) error
}

// StackLister for listing stacks
type StackLister interface {
	ListStacks(stackType StackType) ([]*Stack, error)
}

// StackGetter for getting stacks
type StackGetter interface {
	GetStack(stackName string) (*Stack, error)
}

// StackDeleter for deleting stacks
type StackDeleter interface {
	DeleteStack(stackName string) error
}

// ImageFinder for finding latest image
type ImageFinder interface {
	FindLatestImageID(namePattern string) (string, error)
}

// StackManager composite of all stack capabilities
type StackManager interface {
	StackUpserter
	StackWaiter
	StackLister
	StackGetter
	StackDeleter
	ImageFinder
}

type cloudformationStackManager struct {
	dryrun bool
	cfnAPI cloudformationiface.CloudFormationAPI
	ec2API ec2iface.EC2API
}

// NewStackManager creates a new StackManager backed by cloudformation
func newStackManager(sess *session.Session, dryrun bool) (StackManager, error) {
	log.Debug("Connecting to CloudFormation service")
	cfnAPI := cloudformation.New(sess)

	log.Debug("Connecting to EC2 service")
	ec2API := ec2.New(sess)

	return &cloudformationStackManager{
		dryrun: dryrun,
		cfnAPI: cfnAPI,
		ec2API: ec2API,
	}, nil

}

func buildStackParameters(parameters map[string]string) []*cloudformation.Parameter {
	stackParameters := make([]*cloudformation.Parameter, 0, len(parameters))
	for key, value := range parameters {
		stackParameters = append(stackParameters,
			&cloudformation.Parameter{
				ParameterKey:     aws.String(key),
				ParameterValue:   aws.String(value),
				UsePreviousValue: aws.Bool(value == ""),
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
		})

	for key, value := range tags {
		if value != "" {
			stackTags = append(stackTags,
				&cloudformation.Tag{
					Key:   aws.String(fmt.Sprintf("mu:%s", key)),
					Value: aws.String(value),
				})
		}
	}
	return stackTags
}

// UpsertStack will create/update the cloudformation stack
func (cfnMgr *cloudformationStackManager) UpsertStack(stackName string, templateBodyReader io.Reader, parameters map[string]string, tags map[string]string) error {
	stack := cfnMgr.AwaitFinalStatus(stackName)

	// delete stack if in rollback status
	if stack != nil && stack.Status == cloudformation.StackStatusRollbackComplete {
		log.Warningf("  Stack '%s' was in '%s' status, deleting...", stackName, stack.Status)
		err := cfnMgr.DeleteStack(stackName)
		if err != nil {
			return err
		}
		stack = cfnMgr.AwaitFinalStatus(stackName)
	}

	// load the template
	templateBodyBytes := new(bytes.Buffer)
	templateBodyBytes.ReadFrom(templateBodyReader)
	templateBody := aws.String(templateBodyBytes.String())

	// stack parameters
	stackParameters := buildStackParameters(parameters)

	// stack tags
	stackTags := buildStackTags(tags)

	// directory to write cfn to
	cfnDirectory := fmt.Sprintf("%s/mu-cloudformation", os.TempDir())

	cfnAPI := cfnMgr.cfnAPI
	if stack == nil || stack.Status == "" {

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

		if cfnMgr.dryrun {
			writeTemplateAndConfig(cfnDirectory, stackName, templateBodyBytes, parameters)
			log.Infof("  DRYRUN: Skipping create of stack named '%s'.  Template and parameters written to '%s'", stackName, cfnDirectory)
			return nil
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
		log.Debugf("  Prior state: %s", stack.Status)
		log.Debugf("  Stack parameters:\n\t%s", stackParameters)
		log.Debugf("  Stack tags:\n\t%s", stackTags)
		params := &cloudformation.UpdateStackInput{
			StackName: aws.String(stackName),
			Capabilities: []*string{
				aws.String(cloudformation.CapabilityCapabilityIam),
			},
			Parameters:   stackParameters,
			TemplateBody: templateBody,

			Tags: stackTags,
		}

		if cfnMgr.dryrun {
			writeTemplateAndConfig(cfnDirectory, stackName, templateBodyBytes, parameters)
			log.Infof("  DRYRUN: Skipping update of stack named '%s'.  Template and parameters written to '%s'", stackName, cfnDirectory)
			return nil
		}

		_, err := cfnAPI.UpdateStack(params)
		log.Debug("  Update stack complete err=%s", err)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ValidationError" && awsErr.Message() == "No updates are to be performed." {
					log.Infof("  No changes for stack '%s'", stackName)
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
func (cfnMgr *cloudformationStackManager) AwaitFinalStatus(stackName string) *Stack {
	cfnAPI := cfnMgr.cfnAPI
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	// initialize Spinner
	var statusSpinner *spinner.Spinner
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		statusSpinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	}

	if statusSpinner != nil {
		statusSpinner.Start()
		defer statusSpinner.Stop()
	}

	var priorEventTime *time.Time
	for {

		resp, err := cfnAPI.DescribeStacks(params)

		if statusSpinner != nil {
			statusSpinner.Stop()
		}
		if err != nil || resp == nil || len(resp.Stacks) != 1 {
			log.Debugf("  Stack doesn't exist ... stack=%s", stackName)
			return nil
		}

		if !strings.HasSuffix(aws.StringValue(resp.Stacks[0].StackStatus), "_IN_PROGRESS") {
			log.Debugf("  Returning final status for stack:%s ... status=%s", stackName, *resp.Stacks[0].StackStatus)
			return buildStack(resp.Stacks[0])
		}

		eventParams := &cloudformation.DescribeStackEventsInput{
			StackName: aws.String(stackName),
		}
		eventResp, err := cfnAPI.DescribeStackEvents(eventParams)
		if err == nil && eventResp != nil {
			numEvents := len(eventResp.StackEvents)
			for i := numEvents - 1; i >= 0; i-- {
				e := eventResp.StackEvents[i]
				if priorEventTime == nil || priorEventTime.Before(aws.TimeValue(e.Timestamp)) {

					status := aws.StringValue(e.ResourceStatus)
					eventMesg := fmt.Sprintf("  %s (%s) %s %s", aws.StringValue(e.LogicalResourceId),
						aws.StringValue(e.ResourceType),
						status,
						aws.StringValue(e.ResourceStatusReason))
					if strings.HasSuffix(status, "_IN_PROGRESS") {
						if statusSpinner != nil {
							statusSpinner.Suffix = eventMesg
						}
						log.Debug(eventMesg)
					} else if strings.HasSuffix(status, "_FAILED") {
						log.Error(eventMesg)
					} else {
						log.Debug(eventMesg)
					}
					priorEventTime = e.Timestamp
				}

			}

		}

		log.Debugf("  Not in final status (%s)...sleeping for 5 seconds", *resp.Stacks[0].StackStatus)
		if statusSpinner != nil {
			statusSpinner.Start()
		}
		time.Sleep(time.Second * 5)
	}
}

func buildStack(stackDetails *cloudformation.Stack) *Stack {
	stack := new(Stack)
	stack.ID = aws.StringValue(stackDetails.StackId)
	stack.Name = aws.StringValue(stackDetails.StackName)
	stack.Status = aws.StringValue(stackDetails.StackStatus)
	stack.StatusReason = aws.StringValue(stackDetails.StackStatusReason)
	if aws.TimeValue(stackDetails.LastUpdatedTime).Unix() > 0 {
		stack.LastUpdateTime = aws.TimeValue(stackDetails.LastUpdatedTime)
	} else {
		stack.LastUpdateTime = aws.TimeValue(stackDetails.CreationTime)
	}
	stack.Tags = make(map[string]string)
	stack.Outputs = make(map[string]string)
	stack.Parameters = make(map[string]string)

	for _, tag := range stackDetails.Tags {
		key := aws.StringValue(tag.Key)
		if strings.HasPrefix(key, "mu:") {
			stack.Tags[key[3:]] = aws.StringValue(tag.Value)
		}
	}

	for _, output := range stackDetails.Outputs {
		stack.Outputs[aws.StringValue(output.OutputKey)] = aws.StringValue(output.OutputValue)
	}

	for _, param := range stackDetails.Parameters {
		stack.Parameters[aws.StringValue(param.ParameterKey)] = aws.StringValue(param.ParameterValue)
	}

	return stack
}

// ListStacks will find mu stacks
func (cfnMgr *cloudformationStackManager) ListStacks(stackType StackType) ([]*Stack, error) {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DescribeStacksInput{}

	var stacks []*Stack

	log.Debugf("Searching for stacks of type '%s'", stackType)

	err := cfnAPI.DescribeStacksPages(params,
		func(page *cloudformation.DescribeStacksOutput, lastPage bool) bool {
			for _, stackDetails := range page.Stacks {
				if cloudformation.StackStatusDeleteComplete == aws.StringValue(stackDetails.StackStatus) {
					continue
				}

				stack := buildStack(stackDetails)

				if stack.Tags["type"] == string(stackType) {
					stacks = append(stacks, stack)
				}
			}

			return true
		})

	if err != nil {
		return nil, err
	}
	return stacks, nil
}

// GetStack get a specific stack
func (cfnMgr *cloudformationStackManager) GetStack(stackName string) (*Stack, error) {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}

	log.Debugf("Searching for stack named '%s'", stackName)

	resp, err := cfnAPI.DescribeStacks(params)
	if err != nil {
		return nil, err
	}
	stack := buildStack(resp.Stacks[0])

	return stack, nil
}

// FindLatestImageID for a given
func (cfnMgr *cloudformationStackManager) FindLatestImageID(namePattern string) (string, error) {
	ec2Api := cfnMgr.ec2API
	resp, err := ec2Api.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String("amazon")},
		Filters: []*ec2.Filter{
			{
				Name: aws.String("name"),
				Values: []*string{
					aws.String(namePattern),
				},
			},
		},
	})

	if err != nil {
		return "", err
	}

	var imageID string
	var imageCreateDate time.Time
	for _, image := range resp.Images {
		createDate, err := time.Parse(time.RFC3339, aws.StringValue(image.CreationDate))
		if err != nil {
			return "", err
		}
		if imageCreateDate.Before(createDate) {
			imageCreateDate = createDate
			imageID = aws.StringValue(image.ImageId)
		}
	}

	if imageID == "" {
		return "", errors.New("Unable to find image")
	}
	log.Debugf("Found latest imageId %s for pattern %s", imageID, namePattern)
	return imageID, nil
}

// DeleteStack delete a specific stack
func (cfnMgr *cloudformationStackManager) DeleteStack(stackName string) error {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DeleteStackInput{StackName: aws.String(stackName)}

	if cfnMgr.dryrun {
		log.Infof("  DRYRUN: Skipping delete of stack named '%s'", stackName)
		return nil
	}

	log.Debugf("Deleting stack named '%s'", stackName)

	_, err := cfnAPI.DeleteStack(params)
	return err
}

func writeTemplateAndConfig(cfnDirectory string, stackName string, templateBodyBytes *bytes.Buffer, parameters map[string]string) error {
	os.MkdirAll(cfnDirectory, 0700)
	templateFile := fmt.Sprintf("%s/template-%s.yml", cfnDirectory, stackName)
	err := ioutil.WriteFile(templateFile, templateBodyBytes.Bytes(), 0600)
	if err != nil {
		return err
	}

	configMap := make(map[string]map[string]string)
	configMap["Parameters"] = parameters
	configBody, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return err
	}
	configFile := fmt.Sprintf("%s/config-%s.json", cfnDirectory, stackName)
	err = ioutil.WriteFile(configFile, configBody, 0600)
	return nil
}
