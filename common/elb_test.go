package common

import (
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockedELB struct {
	mock.Mock
	elbv2iface.ELBV2API
}

func (m *mockedELB) DescribeRules(input *elbv2.DescribeRulesInput) (*elbv2.DescribeRulesOutput, error) {
	args := m.Called()
	return args.Get(0).(*elbv2.DescribeRulesOutput), args.Error(1)
}

func TestElbv2Manager_ListRules(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedELB)
	m.On("DescribeRules").Return(
		&elbv2.DescribeRulesOutput{
			Rules: []*elbv2.Rule{},
		}, nil)

	elbManager := elbv2Manager{
		elbAPI: m,
	}

	states, err := elbManager.ListRules("foo")
	assert.Nil(err)
	assert.NotNil(states)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "DescribeRules", 1)
}
