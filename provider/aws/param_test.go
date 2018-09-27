package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedSSM struct {
	mock.Mock
	ssmiface.SSMAPI
}

func (m *mockedSSM) PutParameter(input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	args := m.Called()
	return nil, args.Error(1)
}
func (m *mockedSSM) GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error) {
	args := m.Called()
	return args.Get(0).(*ssm.GetParametersOutput), args.Error(1)
}
func (m *mockedSSM) DescribeParameters(input *ssm.DescribeParametersInput) (*ssm.DescribeParametersOutput, error) {
	args := m.Called()
	return args.Get(0).(*ssm.DescribeParametersOutput), args.Error(1)
}

func TestParamManager_GetParam(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedSSM)
	m.On("GetParameters").Return(&ssm.GetParametersOutput{Parameters: []*ssm.Parameter{{Name: aws.String("foo"), Value: aws.String("bar")}}}, nil)

	paramMgr := paramManager{
		ssmAPI: m,
	}

	val, err := paramMgr.GetParam("foo")
	assert.Nil(err)
	assert.Equal("bar", val)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "GetParameters", 1)
}

func TestParamManager_SetParam(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedSSM)
	m.On("PutParameter").Return(nil, nil)

	paramMgr := paramManager{
		ssmAPI: m,
	}

	err := paramMgr.SetParam("foo", "bar", "key")
	assert.Nil(err)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "PutParameter", 1)
}

func TestParamManager_ParamVersion(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedSSM)
	m.On("DescribeParameters").Return(
		&ssm.DescribeParametersOutput{
			Parameters: []*ssm.ParameterMetadata{
				{
					Name:    aws.String("foo"),
					Version: aws.Int64(2),
				},
			},
			NextToken: aws.String("bar"),
		}, nil)

	paramMgr := paramManager{
		ssmAPI: m,
	}

	val, err := paramMgr.ParamVersion("foo")
	assert.Nil(err)
	assert.Equal(int64(2), val)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "DescribeParameters", 1)
}

func TestParamManager_ParamVersionFalse(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedSSM)
	m.On("DescribeParameters").Return(
		&ssm.DescribeParametersOutput{
			Parameters: []*ssm.ParameterMetadata{},
			NextToken:  aws.String("foo"),
		}, nil)

	paramMgr := paramManager{
		ssmAPI: m,
	}

	val, err := paramMgr.ParamVersion("foo")
	assert.Nil(err)
	assert.Equal(int64(0), val)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "DescribeParameters", 1)
}
