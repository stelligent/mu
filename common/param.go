package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// ParamSetter for setting parameters
type ParamSetter interface {
	SetParam(name string, value string) error
}

// ParamGetter for getting parameters
type ParamGetter interface {
	GetParam(name string) (string, error)
}

// ParamManager composite of all param capabilities
type ParamManager interface {
	ParamGetter
	ParamSetter
}

type paramManager struct {
	ssmAPI ssmiface.SSMAPI
}

func newParamManager(sess *session.Session) (ParamManager, error) {
	log.Debug("Connecting to SSM service")
	ssmAPI := ssm.New(sess)

	return &paramManager{
		ssmAPI: ssmAPI,
	}, nil
}

// SetParam set the value of a parameter
func (paramMgr *paramManager) SetParam(name string, value string) error {
	ssmAPI := paramMgr.ssmAPI

	log.Debug("Setting param '%s' to '%s'", name, value)

	input := &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Type:      aws.String(ssm.ParameterTypeSecureString),
		Overwrite: aws.Bool(true),
	}

	_, err := ssmAPI.PutParameter(input)

	if err != nil {
		return err
	}

	return nil
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
