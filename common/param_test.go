package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
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

	err := paramMgr.SetParam("foo", "bar")
	assert.Nil(err)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "PutParameter", 1)
}
