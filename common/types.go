package common

import (
	"io"
	"time"
)

// Context defines the context object passed around
type Context struct {
	Config          Config
	StackManager    StackManager
	ClusterManager  ClusterManager
	ElbManager      ElbManager
	PipelineManager PipelineManager
	DockerManager   DockerManager
	DockerOut       io.Writer
}

// Config defines the structure of the yml file for the mu config
type Config struct {
	Environments []Environment
	Service      Service
	Basedir      string
	Repo         struct {
		Name     string
		Revision string
	}
}

// Environment defines the structure of the yml file for an environment
type Environment struct {
	Name         string
	Loadbalancer struct {
		HostedZone  string `yaml:"hostedzone"`
		Name        string `yaml:"name"`
		Certificate string `yaml:"certificate"`
		Internal    bool   `yaml:"internal"`
	}
	Cluster struct {
		ImageID           string `yaml:"imageId"`
		InstanceTenancy   string `yaml:"instanceTenancy"`
		DesiredCapacity   int    `yaml:"desiredCapacity"`
		MaxSize           int    `yaml:"maxSize"`
		KeyName           string `yaml:"keyName"`
		SSHAllow          string `yaml:"sshAllow"`
		ScaleOutThreshold int    `yaml:"scaleOutThreshold"`
		ScaleInThreshold  int    `yaml:"scaleInThreshold"`
		HTTPProxy         string `yaml:"httpProxy"`
	}
	VpcTarget struct {
		VpcID        string   `yaml:"vpcId"`
		EcsSubnetIds []string `yaml:"ecsSubnetIds"`
		ElbSubnetIds []string `yaml:"elbSubnetIds"`
	} `yaml:"vpcTarget,omitempty"`
}

// Service defines the structure of the yml file for a service
type Service struct {
	Name            string   `yaml:"name"`
	DesiredCount    int      `yaml:"desiredCount"`
	Dockerfile      string   `yaml:"dockerfile"`
	ImageRepository string   `yaml:"imageRepository"`
	Port            int      `yaml:"port"`
	HealthEndpoint  string   `yaml:"healthEndpoint"`
	CPU             int      `yaml:"cpu"`
	Memory          int      `yaml:"memory"`
	Environment	map[string]string `yaml:"environment"`
	PathPatterns    []string `yaml:"pathPatterns"`
	Priority        int      `yaml:"priority"`
	Pipeline        Pipeline
}

// Pipeline definition
type Pipeline struct {
	Source struct {
		Repo   string `yaml:"repo"`
		Branch string `yaml:"branch"`
	}
	Build struct {
		Type        string `yaml:"type"`
		ComputeType string `yaml:"computeType"`
		Image       string `yaml:"image"`
	}
	Acceptance struct {
		Environment string `yaml:"environment"`
	}
	Production struct {
		Environment string `yaml:"environment"`
	}
	MuBaseurl string `yaml:"muBaseurl"`
	MuVersion string `yaml:"muVersion"`
}

// Stack summary
type Stack struct {
	ID             string
	Name           string
	Status         string
	StatusReason   string
	LastUpdateTime time.Time
	Tags           map[string]string
	Outputs        map[string]string
	Parameters     map[string]string
}

// StackType describes supported stack types
type StackType string

// List of valid stack types
const (
	StackTypeVpc      StackType = "vpc"
	StackTypeTarget             = "target"
	StackTypeCluster            = "cluster"
	StackTypeRepo               = "repo"
	StackTypeService            = "service"
	StackTypePipeline           = "pipeline"
	StackTypeBucket             = "bucket"
)
