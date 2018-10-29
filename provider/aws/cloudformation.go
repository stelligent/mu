package aws

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/briandowns/spinner"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"golang.org/x/crypto/ssh/terminal"
)

type cloudformationStackManager struct {
	dryrunPath        string
	skipVersionCheck  bool
	cfnAPI            cloudformationiface.CloudFormationAPI
	ec2API            ec2iface.EC2API
	extensionsManager common.ExtensionsManager
	statusSpinner     *spinner.Spinner
	spinnerRefCnt     int
	allowDataLoss     bool
}

// NewStackManager creates a new StackManager backed by cloudformation
func newStackManager(sess *session.Session, extensionsManager common.ExtensionsManager, dryrunPath string, skipVersionCheck bool, allowDataLoss bool) (common.StackManager, error) {
	if dryrunPath != "" {
		log.Debugf("Running in DRYRUN mode with path '%s'", dryrunPath)
	}
	log.Debug("Connecting to CloudFormation service")
	cfnAPI := cloudformation.New(sess)

	log.Debug("Connecting to EC2 service")
	ec2API := ec2.New(sess)

	// initialize Spinner
	var statusSpinner *spinner.Spinner
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		statusSpinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	}

	return &cloudformationStackManager{
		dryrunPath:        dryrunPath,
		skipVersionCheck:  skipVersionCheck,
		cfnAPI:            cfnAPI,
		ec2API:            ec2API,
		extensionsManager: extensionsManager,
		statusSpinner:     statusSpinner,
		allowDataLoss:     allowDataLoss,
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
			Value: aws.String(common.GetVersion()),
		})

	for key, value := range tags {
		if value != "" {
			stackTags = append(stackTags,
				&cloudformation.Tag{
					Key:   aws.String(key),
					Value: aws.String(value),
				})
		}
	}
	return stackTags
}

func cleanParams(paramsI interface{}, roleArn string, tags map[string]string) {
	// Reflection to work with both CreateStackInput and UpdateStackInput
	params := reflect.ValueOf(paramsI).Elem()
	roleArnField := params.FieldByName("RoleARN")
	capabilitiesField := params.FieldByName("Capabilities")
	capabilitiesValue := []*string{
		aws.String(cloudformation.CapabilityCapabilityNamedIam),
	}

	if roleArn != "" {
		roleArnField.Set(reflect.ValueOf(aws.String(roleArn)))
	}

	if tags["mu:type"] == string(common.StackTypeIam) {
		capabilitiesField.Set(reflect.ValueOf(capabilitiesValue))
	}
}

func dryrunWrite(cfnMgr *cloudformationStackManager, stackName string, templateBodyBytes *bytes.Buffer, parameters map[string]string,
	operation string) (bool, error) {
	if cfnMgr.dryrunPath != "" {
		err := writeTemplateAndConfig(cfnMgr.dryrunPath, stackName, templateBodyBytes, parameters)
		if err != nil {
			return true, err
		}
		log.Infof("  DRYRUN: Skipping %s of stack named '%s'.  Template and parameters written to '%s'", operation, stackName, cfnMgr.dryrunPath)
		return true, nil
	}
	return false, nil
}

func createStack(stackName string, stackParameters []*cloudformation.Parameter,
	parameters map[string]string,
	roleArn string, stackTags []*cloudformation.Tag, tags map[string]string,
	templateBody *string, templateBodyBytes *bytes.Buffer, policy string,
	cfnMgr *cloudformationStackManager) error {
	operation := "create"
	log.Debugf("  Creating stack named '%s'", stackName)
	log.Debugf("  Stack parameters:\n\t%s", stackParameters)
	log.Debugf("  Assume role:\n\t%s", roleArn)
	log.Debugf("  Stack tags:\n\t%s", stackTags)
	params := &cloudformation.CreateStackInput{
		StackName:        aws.String(stackName),
		Parameters:       stackParameters,
		TemplateBody:     templateBody,
		Tags:             stackTags,
		TimeoutInMinutes: aws.Int64(60),
	}
	if policy != "" {
		params.StackPolicyBody = aws.String(policy)
	}
	cleanParams(params, roleArn, tags)

	dryrun, err := dryrunWrite(cfnMgr, stackName, templateBodyBytes, parameters, operation)
	if err != nil {
		return err
	}
	if dryrun {
		return nil
	}
	log.Debugf("about to cfnAPI.CreateStack(params) with: %v", params)
	_, err = cfnMgr.cfnAPI.CreateStack(params)
	log.Debug("  Create stack complete err=%s", err)
	if err != nil {
		return err
	}
	waitParams := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}
	log.Debug("  Waiting for stack to exist...")
	err = cfnMgr.cfnAPI.WaitUntilStackExists(waitParams)
	if err != nil {
		return err
	}
	log.Debug("  Stack exists.")
	cfnMgr.logInfo("  Created stack '%s'", stackName)
	return nil
}

func updateStack(stackName string, stackParameters []*cloudformation.Parameter,
	parameters map[string]string,
	roleArn string, stackTags []*cloudformation.Tag, tags map[string]string,
	stack *common.Stack, templateBody *string, templateBodyBytes *bytes.Buffer,
	policy string, cfnMgr *cloudformationStackManager) error {
	operation := "update"
	log.Debugf("  Updating stack named '%s'", stackName)
	log.Debugf("  Prior state: %s", stack.Status)
	log.Debugf("  Stack parameters:\n\t%s", stackParameters)
	log.Debugf("  Stack tags:\n\t%s", stackTags)
	params := &cloudformation.UpdateStackInput{
		StackName:    aws.String(stackName),
		Parameters:   stackParameters,
		TemplateBody: templateBody,
		Tags:         stackTags,
	}

	if policy, err := getPolicy(policy, cfnMgr.allowDataLoss); err != nil || policy != "" {
		if err != nil {
			return err
		}
		params.StackPolicyDuringUpdateBody = aws.String(policy)
	}
	cleanParams(params, roleArn, tags)
	dryrun, err := dryrunWrite(cfnMgr, stackName, templateBodyBytes, parameters, operation)
	if err != nil {
		return err
	}
	if dryrun {
		return nil
	}
	_, err = cfnMgr.cfnAPI.UpdateStack(params)
	log.Debug("  Update stack complete err=%s", err)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ValidationError" && awsErr.Message() == "No updates are to be performed." {
				cfnMgr.logInfo("  No changes for stack '%s'", stackName)
				return nil
			}
		}
		return err
	}
	cfnMgr.logInfo("  Updated stack '%s'", stackName)
	return nil
}

func (cfnMgr *cloudformationStackManager) logInfo(format string, args ...interface{}) {
	pauseSpinner := cfnMgr.spinnerRefCnt > 0
	if pauseSpinner {
		cfnMgr.stopSpinner()
	}
	log.Infof(format, args...)
	if pauseSpinner {
		cfnMgr.startSpinner()
	}
}

func checkVersion(cfnMgr *cloudformationStackManager, stack *common.Stack, stackName string) error {
	// check if stack is incompatible
	if !cfnMgr.skipVersionCheck && stack != nil {
		oldMajorVersion, e1 := strconv.Atoi(strings.Split(strings.Split(stack.Tags["version"], "-")[0], ".")[0])
		newMajorVersion, e2 := strconv.Atoi(strings.Split(strings.Split(common.GetVersion(), "-")[0], ".")[0])

		if e1 != nil {
			log.Warningf("Unable to parse major number for existing stack: %s", stack.Tags["version"])
		}

		if e2 != nil {
			log.Warningf("Unable to parse major number for mu: %s", common.GetVersion())
		}

		log.Debugf("comparing stack versions old:%d new:%d", oldMajorVersion, newMajorVersion)

		if e1 == nil && e2 == nil {
			if oldMajorVersion < newMajorVersion {
				return fmt.Errorf("Unable to upsert stack '%s' with existing version '%s' to newer version '%s' (can be overridden with -F)", stackName, stack.Tags["version"], common.GetVersion())
			}
			if oldMajorVersion > newMajorVersion {
				return fmt.Errorf("Unable to upsert stack '%s' with existing version '%s' to older version '%s' (can be overridden with -F)", stackName, stack.Tags["version"], common.GetVersion())
			}
		}
	}
	return nil
}

// cleanStackIfInRollback will remove a given CloudFormation stack (by stackName)
// if it is in ROLLBACK_COMPLETE state.
func (cfnMgr *cloudformationStackManager) cleanStackIfInRollback(stack *common.Stack,
	stackName string) (*common.Stack, error) {
	// delete stack if in rollback status
	if stack != nil && stack.Status == cloudformation.StackStatusRollbackComplete {
		log.Warningf("  Stack '%s' was in '%s' status, deleting...", stackName, stack.Status)
		err := cfnMgr.DeleteStack(stackName)
		if err != nil {
			return nil, err
		}
		stack = cfnMgr.AwaitFinalStatus(stackName)
	}
	return stack, nil
}

func (cfnMgr *cloudformationStackManager) AllowDataLoss(allow bool) {
	cfnMgr.allowDataLoss = allow
}

// SetTerminationProtection to protect stack from deletion
func (cfnMgr *cloudformationStackManager) SetTerminationProtection(stackName string, enabled bool) error {
	if cfnMgr.dryrunPath != "" {
		return nil
	}

	stack := cfnMgr.AwaitFinalStatus(stackName)
	if stack == nil || stack.EnableTerminationProtection == enabled {
		return nil
	}

	if enabled == false && !cfnMgr.allowDataLoss {
		return fmt.Errorf("In order to disable termination protection, you must pass `--allow-data-loss` flag")
	}

	_, err := cfnMgr.cfnAPI.UpdateTerminationProtection(&cloudformation.UpdateTerminationProtectionInput{
		StackName:                   aws.String(stackName),
		EnableTerminationProtection: aws.Bool(enabled),
	})

	return err
}

// UpsertStack will create/update the cloudformation stack
func (cfnMgr *cloudformationStackManager) UpsertStack(stackName string, templateName string, templateData interface{}, parameters map[string]string, tags map[string]string, policy string, roleArn string) error {
	stack := cfnMgr.AwaitFinalStatus(stackName)

	stack, err := cfnMgr.cleanStackIfInRollback(stack, stackName)
	if err != nil {
		return err
	}

	err = checkVersion(cfnMgr, stack, stackName)
	if err != nil {
		return err
	}

	// load the template
	templateBody, err := templates.GetAsset(templateName, templates.ExecuteTemplate(templateData),
		templates.DecorateTemplate(cfnMgr.extensionsManager, stackName))
	if err != nil {
		return err
	}
	templateBodyBytes := bytes.NewBufferString(templateBody)

	// stack parameters
	parameters, err = cfnMgr.extensionsManager.DecorateStackParameters(stackName, parameters)
	if err != nil {
		return err
	}
	stackParameters := buildStackParameters(parameters)

	// stack tags
	tags, err = cfnMgr.extensionsManager.DecorateStackTags(stackName, tags)
	if err != nil {
		return err
	}
	stackTags := buildStackTags(tags)

	if stack == nil || stack.Status == "" {
		// Stack should be created
		err := createStack(stackName, stackParameters, parameters,
			roleArn, stackTags, tags, aws.String(templateBody), templateBodyBytes, policy, cfnMgr)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != "AlreadyExistsException" {
				return err
			}
			log.Debugf("Stack '%s' already existed, trying update instead.", stackName)
			waitParams := &cloudformation.DescribeStacksInput{
				StackName: aws.String(stackName),
			}
			log.Debug("  Waiting for stack to exist...")
			err = cfnMgr.cfnAPI.WaitUntilStackExists(waitParams)
			if err != nil {
				return err
			}
			stack = cfnMgr.AwaitFinalStatus(stackName)
		} else {
			return nil
		}
	}
	// else, stack should be updated
	return updateStack(stackName, stackParameters, parameters,
		roleArn, stackTags, tags, stack, aws.String(templateBody), templateBodyBytes, policy, cfnMgr)
}

func (cfnMgr *cloudformationStackManager) startSpinner() {
	if cfnMgr.statusSpinner != nil {
		cfnMgr.statusSpinner.Start()
		cfnMgr.spinnerRefCnt++
	}
}
func (cfnMgr *cloudformationStackManager) stopSpinner() {
	if cfnMgr.statusSpinner != nil {
		cfnMgr.spinnerRefCnt--
		//if cfnMgr.spinnerRefCnt == 0 {
		cfnMgr.statusSpinner.Stop()
		//}
	}
}

func getPolicy(policy string, allowDataLoss bool) (string, error) {
	if allowDataLoss {
		return templates.GetAsset(common.TemplatePolicyAllowAll)
	}
	return policy, nil
}

func (cfnMgr *cloudformationStackManager) logEventStatus(status string, eventMesg string) {
	if strings.HasSuffix(status, "_IN_PROGRESS") {
		if cfnMgr.statusSpinner != nil {
			cfnMgr.statusSpinner.Suffix = eventMesg
			log.Debug(eventMesg)
		} else {
			log.Info(eventMesg)
		}
	} else if strings.HasSuffix(status, "_FAILED") {
		log.Error(eventMesg)
	} else {
		log.Debug(eventMesg)
	}
}

func (cfnMgr *cloudformationStackManager) logStackEvents(stackName string, priorEventTime *time.Time) {
	cfnAPI := cfnMgr.cfnAPI
	eventParams := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackName),
	}
	eventResp, err := cfnAPI.DescribeStackEvents(eventParams)
	numEvents := len(eventResp.StackEvents)
	if err == nil && eventResp != nil && numEvents > 0 {
		firstEventIndex := 0
		for i := 0; i < numEvents; i++ {
			e := eventResp.StackEvents[i]
			firstEventIndex = i
			if aws.StringValue(e.ResourceType) == "AWS::CloudFormation::Stack" && strings.HasSuffix(aws.StringValue(e.ResourceStatus), "_COMPLETE") {
				break
			}
		}
		for i := firstEventIndex; i >= 0; i-- {
			e := eventResp.StackEvents[i]
			if priorEventTime == nil || priorEventTime.Before(aws.TimeValue(e.Timestamp)) {
				status := aws.StringValue(e.ResourceStatus)
				eventMesg := fmt.Sprintf("  %s:  %s (%s) %s %s",
					stackName,
					aws.StringValue(e.LogicalResourceId),
					aws.StringValue(e.ResourceType),
					status,
					aws.StringValue(e.ResourceStatusReason))
				cfnMgr.logEventStatus(status, eventMesg)

				priorEventTime = e.Timestamp
			}

		}

	}
}

// AwaitFinalStatus waits for the stack to arrive in a final status
//  returns: final status, or empty string if stack doesn't exist
func (cfnMgr *cloudformationStackManager) AwaitFinalStatus(stackName string) *common.Stack {

	cfnAPI := cfnMgr.cfnAPI
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	cfnMgr.startSpinner()
	defer cfnMgr.stopSpinner()

	var priorEventTime *time.Time
	for {

		resp, err := cfnAPI.DescribeStacks(params)

		//cfnMgr.stopSpinner()
		if err != nil || resp == nil || len(resp.Stacks) != 1 {
			// check for stack by a different name (e.g. service catalog)
			stack, err := cfnMgr.GetStack(stackName)
			if err == nil && stack != nil {
				stackName = stack.Name
				params.StackName = aws.String(stackName)
				continue
			}

			log.Debugf("  Stack doesn't exist ... stack=%s", stackName)
			if cfnMgr.dryrunPath != "" {
				stack := &common.Stack{
					Name:           stackName,
					ID:             "",
					Status:         "DRYRUN_COMPLETE",
					StatusReason:   "",
					LastUpdateTime: time.Now(),
					Tags:           make(map[string]string),
					Outputs:        make(map[string]string),
					Parameters:     make(map[string]string),
				}

				stack.Tags["version"] = common.GetVersion()
				log.Debugf("  DRYRUN: Unable to find stack '%s'...returning stub", stackName)
				return stack
			}

			return nil
		}

		if !strings.HasSuffix(aws.StringValue(resp.Stacks[0].StackStatus), "_IN_PROGRESS") {
			log.Debugf("  Returning final status for stack:%s ... status=%s", stackName, *resp.Stacks[0].StackStatus)
			return buildStack(resp.Stacks[0])
		}

		cfnMgr.logStackEvents(stackName, priorEventTime)

		log.Debugf("  Not in final status (%s)...sleeping for 5 seconds", *resp.Stacks[0].StackStatus)
		cfnMgr.startSpinner()
		time.Sleep(time.Second * 5)
	}
}

func buildStack(stackDetails *cloudformation.Stack) *common.Stack {
	stack := new(common.Stack)
	stack.ID = aws.StringValue(stackDetails.StackId)
	stack.Name = aws.StringValue(stackDetails.StackName)
	stack.EnableTerminationProtection = aws.BoolValue(stackDetails.EnableTerminationProtection)
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
func (cfnMgr *cloudformationStackManager) ListStacks(stackType common.StackType, namespace string) ([]*common.Stack, error) {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DescribeStacksInput{}

	var stacks []*common.Stack

	log.Debugf("Searching for stacks of type '%s' in namespace '%s'", stackType, namespace)

	err := cfnAPI.DescribeStacksPages(params,
		func(page *cloudformation.DescribeStacksOutput, lastPage bool) bool {
			for _, stackDetails := range page.Stacks {
				if cloudformation.StackStatusDeleteComplete == aws.StringValue(stackDetails.StackStatus) {
					continue
				}

				stack := buildStack(stackDetails)
				expectedStackPrefix := fmt.Sprintf("%s-%s", namespace, stackType)

				if stack.Tags["type"] == string(stackType) &&
					(strings.HasPrefix(stack.Name, expectedStackPrefix) || strings.HasPrefix(stack.Tags["name"], expectedStackPrefix)) {
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
func (cfnMgr *cloudformationStackManager) GetStack(stackName string) (*common.Stack, error) {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}

	log.Debugf("Searching for stack named '%s'", stackName)

	var stack *common.Stack

	resp, err := cfnAPI.DescribeStacks(params)
	if err == nil {
		stack = buildStack(resp.Stacks[0])
	} else {
		// try finding stack by stack tag 'mu:name'
		err2 := cfnAPI.DescribeStacksPages(&cloudformation.DescribeStacksInput{},
			func(page *cloudformation.DescribeStacksOutput, lastPage bool) bool {
				for _, stackDetails := range page.Stacks {
					if cloudformation.StackStatusDeleteComplete == aws.StringValue(stackDetails.StackStatus) {
						continue
					}

					stack = buildStack(stackDetails)
					if stack.Tags["name"] == stackName {
						err = nil
						return false
					}
				}

				return true
			})
		if err2 != nil {
			return nil, err2
		}
		return stack, err
	}
	return stack, nil
}

// CountAZs for current region
func (cfnMgr *cloudformationStackManager) CountAZs() (int, error) {
	ec2Api := cfnMgr.ec2API
	resp, err := ec2Api.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("state"),
				Values: []*string{
					aws.String("available"),
				},
			},
		},
	})
	if err != nil {
		return 0, err
	}
	return len(resp.AvailabilityZones), nil
}

// FindLatestImageID for a given
func (cfnMgr *cloudformationStackManager) FindLatestImageID(owner string, namePattern string) (string, error) {
	ec2Api := cfnMgr.ec2API
	resp, err := ec2Api.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String(owner)},
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

	if cfnMgr.dryrunPath != "" {
		log.Infof("  DRYRUN: Skipping delete of stack named '%s'", stackName)
		return nil
	}

	log.Debugf("Deleting stack named '%s'", stackName)

	_, err := cfnAPI.DeleteStack(params)
	if err != nil {
		return err
	}
	cfnMgr.logInfo("  Deleted stack '%s'", stackName)
	return nil
}

func writeTemplateAndConfig(cfnDirectory string, stackName string, templateBodyBytes *bytes.Buffer, parameters map[string]string) error {
	err := os.MkdirAll(cfnDirectory, 0700)
	if err != nil {
		return err
	}
	templateFile := fmt.Sprintf("%s/template-%s.yml", cfnDirectory, stackName)
	err = ioutil.WriteFile(templateFile, templateBodyBytes.Bytes(), 0600)
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
	return ioutil.WriteFile(configFile, configBody, 0600)
}
