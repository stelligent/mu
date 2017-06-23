package common

import (
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

// RdsIamAuthenticationSetter for getting db instances
type RdsIamAuthenticationSetter interface {
	SetIamAuthentication(dbInstanceIdentifier string, enabled bool, dbEngine string) error
}

// RdsManager composite of all cluster capabilities
type RdsManager interface {
	RdsIamAuthenticationSetter
}

type rdsManager struct {
	rdsAPI rdsiface.RDSAPI
}
