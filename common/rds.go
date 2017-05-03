package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

// RdsIamAuthenticationSetter for getting cluster instances
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

func newRdsManager(sess *session.Session) (RdsManager, error) {
	log.Debug("Connecting to RDS service")
	rdsAPI := rds.New(sess)

	return &rdsManager{
		rdsAPI: rdsAPI,
	}, nil
}

// ListState get the state of the pipeline
func (rdsMgr *rdsManager) SetIamAuthentication(dbInstanceIdentifier string, enabled bool, dbEngine string) error {
	rdsAPI := rdsMgr.rdsAPI

	var err error
	if dbEngine == "aurora" || dbEngine == "" {
		params := &rds.ModifyDBClusterInput{
			DBClusterIdentifier:             aws.String(dbInstanceIdentifier),
			EnableIAMDatabaseAuthentication: aws.Bool(true),
			ApplyImmediately:                aws.Bool(true),
		}

		log.Debugf("Setting IAM Authentication to '%s' for '%s'", enabled, dbInstanceIdentifier)

		_, err = rdsAPI.ModifyDBCluster(params)
	} else {
		params := &rds.ModifyDBInstanceInput{
			DBInstanceIdentifier:            aws.String(dbInstanceIdentifier),
			EnableIAMDatabaseAuthentication: aws.Bool(true),
			ApplyImmediately:                aws.Bool(true),
		}

		log.Debugf("Setting IAM Authentication to '%s' for '%s'", enabled, dbInstanceIdentifier)

		_, err = rdsAPI.ModifyDBInstance(params)
	}

	if err != nil {
		return err
	}

	return nil
}
