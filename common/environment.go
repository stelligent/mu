package common

import (
	"fmt"
)

// Environment defines the env that will be created
type Environment struct {
	Name string `yaml:"name"`
	Loadbalancer EnvironmentLoadBalancer `yaml:"loadbalancer,omitempty"`
	Cluster EnvironmentCluster `yaml:"cluster,omitempty"`
}

// EnvironmentLoadBalancer defines the env load balancer that will be created
type EnvironmentLoadBalancer struct {
	Hostname string `yaml:"hostname,omitempty"`
}

// EnvironmentCluster defines the env cluster that will be created
type EnvironmentCluster struct {
	DesiredCapacity int `yaml:"desiredCapacity,omitempty"`
	MaxSize int `yaml:"maxSize,omitempty"`
}

// NewStack will create a new stack instance and write the template for the stack
func (environment *Environment) NewStack() (*Stack, error) {
	stackName := fmt.Sprintf("mu-env-%s", environment.Name)
	stack := NewStack(stackName)

	err := stack.WriteTemplate("environment-template.yml", environment)
	if err != nil {
		return nil, err
	}

	return stack,nil
}
