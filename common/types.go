package common

// Context defines the context object passed around
type Context struct {
	Config       Config
	StackManager StackManager
}

// Config defines the structure of the yml file for the mu config
type Config struct {
	Region       string
	Environments []Environment
	Service      Service
}

// Environment defines the structure of the yml file for an environment
type Environment struct {
	Name         string
	Loadbalancer struct {
		Hostname string
	}
	Cluster struct {
		DesiredCapacity int `yaml:"desiredCapacity"`
		MaxSize         int `yaml:"maxSize"`
	}
	VpcTarget struct {
		VpcID           string   `yaml:"vpcId"`
		PublicSubnetIds []string `yaml:"publicSubnetIds"`
	} `yaml:"vpcTarget,omitempty"`
}

// Service defines the structure of the yml file for a service
type Service struct {
	DesiredCount int `yaml:"desiredCount"`
	Pipeline     struct {
	}
}

// Stack summary
type Stack struct {
	ID           string
	Name         string
	Status       string
	StatusReason string
	Tags         map[string]string
}

// StackType describes supported stack types
type StackType string

// List of valid stack types
const (
	StackTypeVpc      StackType = "vpc"
	StackTypeCluster            = "cluster"
	StackTypeService            = "service"
	StackTypePipeline           = "pipeline"
)
