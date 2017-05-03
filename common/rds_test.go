package common

import (
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockedRDS struct {
	mock.Mock
	rdsiface.RDSAPI
}

func (m *mockedRDS) ModifyDBInstance(input *rds.ModifyDBInstanceInput) (*rds.ModifyDBInstanceOutput, error) {
	args := m.Called()
	return nil, args.Error(1)
}

func TestRdsManager_SetIamAuthentication(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedRDS)
	m.On("ModifyDBInstance").Return(nil, nil)

	rds := rdsManager{
		rdsAPI: m,
	}

	err := rds.SetIamAuthentication("foo", true, "postgres")
	assert.Nil(err)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "ModifyDBInstance", 1)
}
