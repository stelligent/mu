package common

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// Context defines the context object passed around
type Context struct {
	Config Config
	CloudFormation cloudformationiface.CloudFormationAPI
}

// Config defines the structure of the yml file for the mu config
type Config struct {
	Region string
	Environments []Environment
	Service Service
}

// Environment defines the structure of the yml file for an environment
type Environment struct {
	Name string
	Loadbalancer struct {
		     Hostname string
	}
	Cluster struct {
		     DesiredCapacity int `yaml:"desiredCapacity"`
		     MaxSize int `yaml:"maxSize"`
	}
}

// Service defines the structure of the yml file for a service
type Service struct {
	DesiredCount int`yaml:"desiredCount"`
	Pipeline struct {
	}
}
