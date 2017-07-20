package common

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

// InstanceLister for getting instances
type InstanceLister interface {
	ListInstances(instanceIds ...string) ([]Instance, error)
}

// Instance represents and EC2 instance
type Instance *ec2.Instance

// InstanceManager composite of all instance capabilities
type InstanceManager interface {
	InstanceLister
}
