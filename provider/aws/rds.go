package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/stelligent/mu/common"
)

type rdsManager struct {
	rdsAPI rdsiface.RDSAPI
}

func newRdsManager(sess *session.Session) (common.RdsManager, error) {
	log.Debug("Connecting to RDS service")
	rdsAPI := rds.New(sess)

	return &rdsManager{
		rdsAPI: rdsAPI,
	}, nil
}

// SetIamAuthentication set value of RDS IAM authentication
func (rdsMgr *rdsManager) SetIamAuthentication(dbInstanceIdentifier string, enabled bool, dbEngine string) error {
	rdsAPI := rdsMgr.rdsAPI

	var err error
	if dbEngine == "aurora" || dbEngine == "aurora-postgresql" || dbEngine == "" {
		params := &rds.ModifyDBClusterInput{
			DBClusterIdentifier:             aws.String(dbInstanceIdentifier),
			EnableIAMDatabaseAuthentication: aws.Bool(enabled),
			ApplyImmediately:                aws.Bool(true),
		}

		log.Debugf("Setting IAM Authentication to '%s' for '%s'", enabled, dbInstanceIdentifier)

		_, err = rdsAPI.ModifyDBCluster(params)
	} else {
		params := &rds.ModifyDBInstanceInput{
			DBInstanceIdentifier:            aws.String(dbInstanceIdentifier),
			EnableIAMDatabaseAuthentication: aws.Bool(enabled),
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
