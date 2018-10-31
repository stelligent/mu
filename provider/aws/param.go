package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stelligent/mu/common"
)

type paramManager struct {
	ssmAPI ssmiface.SSMAPI
	dryrun bool
}

func newParamManager(sess *session.Session, dryrun bool) (common.ParamManager, error) {
	log.Debug("Connecting to SSM service")
	ssmAPI := ssm.New(sess)

	return &paramManager{
		dryrun: dryrun,
		ssmAPI: ssmAPI,
	}, nil
}

// SetParam set the value of a parameter
func (paramMgr *paramManager) SetParam(name string, value string, kmsKey string) error {
	ssmAPI := paramMgr.ssmAPI

	if paramMgr.dryrun {
		return nil
	}

	log.Debug("Setting param '%s' to '%s'", name, value)

	input := &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Type:      aws.String(ssm.ParameterTypeSecureString),
		KeyId:     aws.String(kmsKey),
		Overwrite: aws.Bool(true),
	}

	_, err := ssmAPI.PutParameter(input)

	if err != nil {
		return err
	}

	return nil
}

// DeleteParam set the value of a parameter
func (paramMgr *paramManager) DeleteParam(name string) error {
	ssmAPI := paramMgr.ssmAPI

	if paramMgr.dryrun {
		return nil
	}

	log.Debug("Deleting param '%s'", name)

	input := &ssm.DeleteParameterInput{
		Name: aws.String(name),
	}

	_, err := ssmAPI.DeleteParameter(input)

	return err
}

// GetParam get the value of a parameter
func (paramMgr *paramManager) GetParam(name string) (string, error) {
	ssmAPI := paramMgr.ssmAPI

	log.Debug("Getting param '%s'", name)

	input := &ssm.GetParametersInput{
		Names:          []*string{aws.String(name)},
		WithDecryption: aws.Bool(true),
	}

	output, err := ssmAPI.GetParameters(input)

	if err != nil {
		return "", err
	}

	if len(output.Parameters) != 1 {
		return "", nil
	}

	return aws.StringValue(output.Parameters[0].Value), nil
}

// ParamVersion checks if the parameter is set in SSM Parameter Store and return version
func (paramMgr *paramManager) ParamVersion(name string) (int64, error) {
	ssmAPI := paramMgr.ssmAPI

	log.Debug("checking param exists '%s'", name)

	input := &ssm.DescribeParametersInput{
		Filters: []*ssm.ParametersFilter{
			{
				Key:    aws.String("Name"),
				Values: []*string{aws.String(name)},
			},
		},
	}

	output, err := ssmAPI.DescribeParameters(input)

	if err != nil {
		return 0, err
	}

	if len(output.Parameters) != 1 {
		return 0, nil
	}

	return *output.Parameters[0].Version, nil
}
