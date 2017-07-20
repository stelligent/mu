package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/stelligent/mu/common"
)

type ec2InstanceManager struct {
	ec2API ec2iface.EC2API
}

func newInstanceManager(sess *session.Session) (common.InstanceManager, error) {
	log.Debug("Connecting to EC2 service")
	ec2API := ec2.New(sess)

	return &ec2InstanceManager{
		ec2API: ec2API,
	}, nil
}

// ListInstances get the instances for a specific cluster
func (ec2Mgr *ec2InstanceManager) ListInstances(instanceIds ...string) ([]common.Instance, error) {
	ec2InputParameters := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice(instanceIds),
	}

	var instances []common.Instance
	ec2Mgr.ec2API.DescribeInstancesPages(ec2InputParameters, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				instances = append(instances, instance)
			}
		}
		return true
	})

	return instances, nil
}
