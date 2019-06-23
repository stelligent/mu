package common

import (
	"fmt"
	"io"
	"time"
)

// Context defines the context object passed around
type Context struct {
	Config                            Config
	AccountID                         string
	Partition                         string
	Region                            string
	StackManager                      StackManager
	ClusterManager                    ClusterManager
	InstanceManager                   InstanceManager
	ElbManager                        ElbManager
	RdsManager                        RdsManager
	ParamManager                      ParamManager
	LocalPipelineManager              PipelineManager // instance that ignores region/profile/role
	PipelineManager                   PipelineManager
	LogsManager                       LogsManager
	DockerManager                     DockerManager
	DockerOut                         io.Writer
	KubernetesResourceManagerProvider KubernetesResourceManagerProvider
	TaskManager                       TaskManager
	ArtifactManager                   ArtifactManager
	SubscriptionManager               SubscriptionManager
	RolesetManager                    RolesetManager
	ExtensionsManager                 ExtensionsManager
	CatalogManager                    CatalogManager
}

// Config defines the structure of the yml file for the mu config
type Config struct {
	DryRun       bool          `yaml:"-"`
	Namespace    string        `yaml:"namespace,omitempty" validate:"validateAlphaNumericDash"`
	Environments []Environment `yaml:"environments,omitempty"`
	Service      Service       `yaml:"service,omitempty"`
	Basedir      string        `yaml:"-"`
	RelMuFile    string        `yaml:"-"`
	Repo         struct {
		Name     string
		Slug     string
		Revision string
		Branch   string
		Provider string
	} `yaml:"-"`
	Templates  map[string]interface{}       `yaml:"templates,omitempty"`
	Parameters map[string]map[string]string `yaml:"parameters,omitempty"`
	Tags       map[string]map[string]string `yaml:"tags,omitempty"`
	Extensions []Extension                  `yaml:"extensions,omitempty"`
	DisableIAM bool                         `yaml:"disableIAM,omitempty"`
	Roles      struct {
		CloudFormation string `yaml:"cloudFormation,omitempty" validate:"validateRoleARN"`
	} `yaml:"roles,omitempty"`
	RBAC    []RoleBinding `yaml:"rbac,omitempty"`
	Catalog Catalog       `yaml:"catalog,omitempty"`
}

// Catalog of pipeline templates
type Catalog struct {
	IAMUsers  []string `yaml:"iamUsers,omitempty"`
	Pipelines []struct {
		Name        string              `yaml:"name,omitempty" validate:"validateAlphaNumericDash"`
		Description string              `yaml:"description,omitempty"`
		Versions    map[string]Pipeline `yaml:"versions,omitempty"`
	} `yaml:"pipelines,omitempty"`
}

// RoleBinding defines how to map k8s roles to subjects
type RoleBinding struct {
	Role         RBACRole `yaml:"role,omitempty"`
	Environments []string `yaml:"environments,omitempty"`
	Users        []string `yaml:"users,omitempty"`
	Services     []string `yaml:"services,omitempty"`
}

// RBACRole describes possible rbac roles
type RBACRole string

// List of valid stack types
const (
	RBACRoleAdmin  RBACRole = "admin"
	RBACRoleView            = "view"
	RBACRoleDeploy          = "deploy"
)

// Extension defines the structure of the yml file for an extension
type Extension struct {
	URL   string `yaml:"url,omitempty"`
	Image string `yaml:"image,omitempty"`
}

// Environment defines the structure of the yml file for an environment
type Environment struct {
	Name         string       `yaml:"name,omitempty" validate:"validateLeadingAlphaNumericDash"`
	Provider     EnvProvider  `yaml:"provider,omitempty"`
	Loadbalancer Loadbalancer `yaml:"loadbalancer,omitempty"`
	Cluster      Cluster      `yaml:"cluster,omitempty"`
	Discovery    struct {
		Provider string `yaml:"provider,omitempty"`
		Name     string `yaml:"name,omitempty"`
	} `yaml:"discovery,omitempty"`
	VpcTarget VpcTarget        `yaml:"vpcTarget,omitempty"`
	Roles     EnvironmentRoles `yaml:"roles,omitempty"`
}

// Loadbalancer defines the scructure of the yml file for a loadbalancer
type Loadbalancer struct {
	HostedZone  string `yaml:"hostedzone,omitempty" validate:"validateURL"`
	Name        string `yaml:"name,omitempty"  validate:"validateLeadingAlphaNumericDash=32"`
	Certificate string `yaml:"certificate,omitempty"`
	Internal    bool   `yaml:"internal,omitempty"`
	AccessLogs  struct {
		S3BucketName string `yaml:"s3BucketName,omitempty"`
		S3Prefix     string `yaml:"s3Prefix,omitempty"`
	} `yaml:"accessLogs,omitempty"`
}

// Cluster defines the scructure of the yml file for a cluster of EC2 instance AWS::AutoScaling::LaunchConfiguration
type Cluster struct {
	InstanceType            string          `yaml:"instanceType,omitempty" validate:"validateInstanceType"`
	ImageID                 string          `yaml:"imageId,omitempty" validate:"validateResourceID=ami"`
	ImageOsType             string          `yaml:"osType,omitempty"`
	InstanceTenancy         InstanceTenancy `yaml:"instanceTenancy,omitempty"`
	DesiredCapacity         int             `yaml:"desiredCapacity,omitempty"`
	MinSize                 int             `yaml:"minSize,omitempty"`
	MaxSize                 int             `yaml:"maxSize,omitempty"`
	KeyName                 string          `yaml:"keyName,omitempty"`
	SSHAllow                string          `yaml:"sshAllow,omitempty" validate:"validateCIDR"`
	TargetCPUReservation    int             `yaml:"targetCPUReservation,omitempty" validate:"max=100"`
	TargetMemoryReservation int             `yaml:"targetMemoryReservation,omitempty" validate:"max=100"`
	HTTPProxy               string          `yaml:"httpProxy,omitempty"  validate:"validateURL"`
	ExtraUserData           string          `yaml:"extraUserData,omitempty"`
}

// VpcTarget defines the structure of the yml file for a cluster VPC
type VpcTarget struct {
	VpcID             string   `yaml:"vpcId,omitempty" validate:"validateResourceID=vpc"`
	InstanceSubnetIds []string `yaml:"instanceSubnetIds,omitempty" validate:"validateResourceID=subnet"`
	ElbSubnetIds      []string `yaml:"elbSubnetIds,omitempty" validate:"validateResourceID=subnet"`
	Environment       string   `yaml:"environment" validate:"validateLeadingAlphaNumericDash"`
	Namespace         string   `yaml:"namespace" validate:"validateLeadingAlphaNumericDash"`
}

// EnvironmentRoles defines the structure of the yml file for environment roles
type EnvironmentRoles struct {
	Instance   string `yaml:"instance,omitempty" validate:"validateRoleARN"`
	EksService string `yaml:"eksService,omitempty" validate:"validateRoleARN"`
}

// Service defines the structure of the yml file for a service
type Service struct {
	Name                 string                 `yaml:"name,omitempty" validate:"validateLeadingAlphaNumericDash"`
	DeploymentStrategy   DeploymentStrategy     `yaml:"deploymentStrategy,omitempty"`
	DesiredCount         int                    `yaml:"desiredCount,omitempty"`
	MinSize              int                    `yaml:"minSize,omitempty"`
	MaxSize              int                    `yaml:"maxSize,omitempty"`
	Dockerfile           string                 `yaml:"dockerfile,omitempty"`
	ImageRepository      string                 `yaml:"imageRepository,omitempty"`
	Port                 int                    `yaml:"port,omitempty" validate:"max=65535"`
	Protocol             ServiceProtocol        `yaml:"protocol,omitempty"`
	HealthEndpoint       string                 `yaml:"healthEndpoint,omitempty" validate:"validateURL"`
	CPU                  int                    `yaml:"cpu,omitempty"`
	Memory               int                    `yaml:"memory,omitempty"`
	NetworkMode          NetworkMode            `yaml:"networkMode,omitempty"`
	AssignPublicIP       bool                   `yaml:"assignPublicIp,omitempty"`
	Links                []string               `yaml:"links,omitempty"`
	Environment          map[string]interface{} `yaml:"environment,omitempty"`
	Secrets              map[string]interface{} `yaml:"secrets,omitempty"`
	PathPatterns         []string               `yaml:"pathPatterns,omitempty"`
	HostPatterns         []string               `yaml:"hostPatterns,omitempty"`
	Priority             int                    `yaml:"priority,omitempty" validate:"max=50000"`
	Pipeline             Pipeline               `yaml:"pipeline,omitempty"`
	Database             Database               `yaml:"database,omitempty"`
	Schedule             []Schedule             `yaml:"schedules,omitempty"`
	TargetCPUUtilization int                    `yaml:"targetCPUUtilization,omitempty" validate:"max=100"`
	DiscoveryTTL         string                 `yaml:"discoveryTTL,omitempty"`
	Roles                struct {
		Ec2Instance            string `yaml:"ec2Instance,omitempty" validate:"validateRoleARN"`
		CodeDeploy             string `yaml:"codeDeploy,omitempty" validate:"validateRoleARN"`
		EcsEvents              string `yaml:"ecsEvents,omitempty" validate:"validateRoleARN"`
		EcsService             string `yaml:"ecsService,omitempty" validate:"validateRoleARN"`
		EcsTask                string `yaml:"ecsTask,omitempty" validate:"validateRoleARN"`
		ApplicationAutoScaling string `yaml:"applicationAutoScaling,omitempty" validate:"validateRoleARN"`
	} `yaml:"roles,omitempty"`
}

// Database definition
type Database struct {
	DatabaseConfig    `yaml:",inline"`
	EnvironmentConfig map[string]DatabaseConfig `yaml:"environmentConfig"`
}

// DatabaseConfig definition
type DatabaseConfig struct {
	Name                   string `yaml:"name,omitempty" validate:"validateLeadingAlphaNumericDash"`
	InstanceClass          string `yaml:"instanceClass,omitempty" validate:"validateInstanceType"`
	Engine                 string `yaml:"engine,omitempty" validate:"validateAlphaNumericDash"`
	EngineMode             string `yaml:"engineMode,omitempty" validate:"validateAlphaNumericDash"`
	IamAuthentication      string `yaml:"iamAuthentication,omitempty"`
	MasterUsername         string `yaml:"masterUsername,omitempty"`
	AllocatedStorage       string `yaml:"allocatedStorage,omitempty"`
	KmsKey                 string `yaml:"kmsKey,omitempty"`
	MinSize                string `yaml:"minSize,omitempty"`
	MaxSize                string `yaml:"maxSize,omitempty"`
	SecondsUntilAutoPause  string `yaml:"secondsUntilAutoPause,omitempty"`
	MasterPasswordSSMParam string `yaml:"masterPasswordSSMParam,omitempty"`
}

// GetDatabaseConfig definition
func (database *Database) GetDatabaseConfig(environmentName string) *DatabaseConfig {
	first := func(options ...string) string {
		for _, s := range options {
			if s != "" {
				return s
			}
		}
		return ""
	}
	envConfig := database.EnvironmentConfig[environmentName]
	dbConfig := &DatabaseConfig{
		Name:                   first(envConfig.Name, database.Name),
		InstanceClass:          first(envConfig.InstanceClass, database.InstanceClass),
		Engine:                 first(envConfig.Engine, database.Engine),
		EngineMode:             first(envConfig.EngineMode, database.EngineMode),
		IamAuthentication:      first(envConfig.IamAuthentication, database.IamAuthentication),
		MasterUsername:         first(envConfig.MasterUsername, database.MasterUsername),
		AllocatedStorage:       first(envConfig.AllocatedStorage, database.AllocatedStorage),
		KmsKey:                 first(envConfig.KmsKey, database.KmsKey),
		MinSize:                first(envConfig.MinSize, database.MinSize),
		MaxSize:                first(envConfig.MaxSize, database.MaxSize),
		SecondsUntilAutoPause:  first(envConfig.SecondsUntilAutoPause, database.SecondsUntilAutoPause),
		MasterPasswordSSMParam: first(envConfig.MasterPasswordSSMParam, database.MasterPasswordSSMParam),
	}
	return dbConfig
}

// Schedule definition
type Schedule struct {
	Name       string   `yaml:"name,omitempty" validate:"validateLeadingAlphaNumericDash"`
	Expression string   `yaml:"expression,omitempty"`
	Command    []string `yaml:"command,omitempty"`
}

// Pipeline definition
type Pipeline struct {
	Catalog struct {
		Name    string `yaml:"name,omitempty"`
		Version string `yaml:"version,omitempty"`
	} `yaml:"catalog,omitempty"`
	Source struct {
		Provider string `yaml:"provider,omitempty"`
		Repo     string `yaml:"repo,omitempty"`
		Branch   string `yaml:"branch,omitempty"`
	} `yaml:"source,omitempty"`
	Build struct {
		Disabled     bool            `yaml:"disabled,omitempty"`
		Type         EnvironmentType `yaml:"type,omitempty"`
		ComputeType  ComputeType     `yaml:"computeType,omitempty"`
		Image        string          `yaml:"image,omitempty" validate:"validateDockerImage"`
		Bucket       string          `yaml:"bucket,omitempty"`
		BuildTimeout string          `yaml:"timeout,omitempty" validate:"max=480"`
	} `yaml:"build,omitempty"`
	Acceptance struct {
		Disabled    bool            `yaml:"disabled,omitempty"`
		Environment string          `yaml:"environment,omitempty"`
		Type        EnvironmentType `yaml:"type,omitempty"`
		ComputeType ComputeType     `yaml:"computeType,omitempty"`
		Image       string          `yaml:"image,omitempty" validate:"validateDockerImage"`
		Roles       struct {
			CodeBuild string `yaml:"codeBuild,omitempty" validate:"validateRoleARN"`
			Mu        string `yaml:"mu,omitempty" validate:"validateRoleARN"`
		} `yaml:"roles,omitempty"`
		BuildTimeout string `yaml:"timeout,omitempty" validate:"max=480"`
	} `yaml:"acceptance,omitempty"`
	Production struct {
		Disabled    bool   `yaml:"disabled,omitempty"`
		Environment string `yaml:"environment,omitempty"`
		Roles       struct {
			CodeBuild string `yaml:"codeBuild,omitempty" validate:"validateRoleARN"`
			Mu        string `yaml:"mu,omitempty" validate:"validateRoleARN"`
		} `yaml:"roles,omitempty"`
		BuildTimeout string `yaml:"timeout,omitempty" validate:"max=480"`
	} `yaml:"production,omitempty"`
	MuBaseurl string `yaml:"muBaseurl,omitempty"`
	MuVersion string `yaml:"muVersion,omitempty"`
	KmsKey    string `yaml:"kmsKey,omitempty"`
	Roles     struct {
		Pipeline string `yaml:"pipeline,omitempty" validate:"validateRoleARN"`
		Build    string `yaml:"build,omitempty" validate:"validateRoleARN"`
	} `yaml:"roles,omitempty"`
	Bucket string   `yaml:"bucket,omitempty"`
	Notify []string `yaml:"notify,omitempty"`
}

// Stack summary
type Stack struct {
	ID                          string
	Name                        string
	EnableTerminationProtection bool
	Status                      string
	StatusReason                string
	LastUpdateTime              time.Time
	Tags                        map[string]string
	Outputs                     map[string]string
	Parameters                  map[string]string
}

const (
	// StackStatusCreateInProgress is a StackStatus enum value
	StackStatusCreateInProgress = "CREATE_IN_PROGRESS"

	// StackStatusCreateFailed is a StackStatus enum value
	StackStatusCreateFailed = "CREATE_FAILED"

	// StackStatusCreateComplete is a StackStatus enum value
	StackStatusCreateComplete = "CREATE_COMPLETE"

	// StackStatusRollbackInProgress is a StackStatus enum value
	StackStatusRollbackInProgress = "ROLLBACK_IN_PROGRESS"

	// StackStatusRollbackFailed is a StackStatus enum value
	StackStatusRollbackFailed = "ROLLBACK_FAILED"

	// StackStatusRollbackComplete is a StackStatus enum value
	StackStatusRollbackComplete = "ROLLBACK_COMPLETE"

	// StackStatusDeleteInProgress is a StackStatus enum value
	StackStatusDeleteInProgress = "DELETE_IN_PROGRESS"

	// StackStatusDeleteFailed is a StackStatus enum value
	StackStatusDeleteFailed = "DELETE_FAILED"

	// StackStatusDeleteComplete is a StackStatus enum value
	StackStatusDeleteComplete = "DELETE_COMPLETE"

	// StackStatusUpdateInProgress is a StackStatus enum value
	StackStatusUpdateInProgress = "UPDATE_IN_PROGRESS"

	// StackStatusUpdateCompleteCleanupInProgress is a StackStatus enum value
	StackStatusUpdateCompleteCleanupInProgress = "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"

	// StackStatusUpdateComplete is a StackStatus enum value
	StackStatusUpdateComplete = "UPDATE_COMPLETE"

	// StackStatusUpdateRollbackInProgress is a StackStatus enum value
	StackStatusUpdateRollbackInProgress = "UPDATE_ROLLBACK_IN_PROGRESS"

	// StackStatusUpdateRollbackFailed is a StackStatus enum value
	StackStatusUpdateRollbackFailed = "UPDATE_ROLLBACK_FAILED"

	// StackStatusUpdateRollbackCompleteCleanupInProgress is a StackStatus enum value
	StackStatusUpdateRollbackCompleteCleanupInProgress = "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"

	// StackStatusUpdateRollbackComplete is a StackStatus enum value
	StackStatusUpdateRollbackComplete = "UPDATE_ROLLBACK_COMPLETE"

	// StackStatusReviewInProgress is a StackStatus enum value
	StackStatusReviewInProgress = "REVIEW_IN_PROGRESS"
)

// StackType describes supported stack types
type StackType string

// List of valid stack types
const (
	StackTypeVpc          StackType = "vpc"
	StackTypeTarget                 = "target"
	StackTypeIam                    = "iam"
	StackTypeEnv                    = "environment"
	StackTypeLoadBalancer           = "loadbalancer"
	StackTypeRepo                   = "repo"
	StackTypeApp                    = "app"
	StackTypeService                = "service"
	StackTypePipeline               = "pipeline"
	StackTypeDatabase               = "database"
	StackTypeSchedule               = "schedule"
	StackTypeBucket                 = "bucket"
	StackTypePortfolio              = "portfolio"
	StackTypeProduct                = "product"
)

// List of valid template files
const (
	TemplatePolicyDefault    string = "policies/default.json"
	TemplatePolicyAllowAll          = "policies/allow-all.json"
	TemplateBuildspec               = "codebuild/buildspec.yml"
	TemplateApp                     = "cloudformation/app.yml"
	TemplateBucket                  = "cloudformation/bucket.yml"
	TemplatePortfolio               = "cloudformation/portfolio.yml"
	TemplatePortfolioIAM            = "cloudformation/portfolio-iam.yml"
	TemplateProduct                 = "cloudformation/product.yml"
	TemplateCommonIAM               = "cloudformation/common-iam.yml"
	TemplateDatabase                = "cloudformation/database.yml"
	TemplateELB                     = "cloudformation/elb.yml"
	TemplateEnvEC2                  = "cloudformation/env-ec2.yml"
	TemplateEnvECS                  = "cloudformation/env-ecs.yml"
	TemplateEnvEKS                  = "cloudformation/env-eks.yml"
	TemplateEnvEKSBootstrap         = "cloudformation/env-eks-bootstrap.yml"
	TemplateEnvIAM                  = "cloudformation/env-iam.yml"
	TemplatePipelineIAM             = "cloudformation/pipeline-iam.yml"
	TemplatePipeline                = "cloudformation/pipeline.yml"
	TemplateRepo                    = "cloudformation/repo.yml"
	TemplateSchedule                = "cloudformation/schedule.yml"
	TemplateServiceEC2              = "cloudformation/service-ec2.yml"
	TemplateServiceECS              = "cloudformation/service-ecs.yml"
	TemplateServiceIAM              = "cloudformation/service-iam.yml"
	TemplateVPCTarget               = "cloudformation/vpc-target.yml"
	TemplateVPC                     = "cloudformation/vpc.yml"
	TemplateK8sCluster              = "kubernetes/cluster.yml"
	TemplateK8sDeployment           = "kubernetes/deployment.yml"
	TemplateK8sDatabase             = "kubernetes/database.yml"
	TemplateK8sIngress              = "kubernetes/ingress.yml"
	TemplateArtifactPipeline        = "cloudformation/artifact-pipeline.yml"
)

// DeploymentStrategy describes supported deployment strategies
type DeploymentStrategy string

// List of supported deployment strategies
const (
	BlueGreenDeploymentStrategy DeploymentStrategy = "blue_green"
	RollingDeploymentStrategy   DeploymentStrategy = "rolling"
	ReplaceDeploymentStrategy   DeploymentStrategy = "replace"
)

// EnvProvider describes supported environment strategies
type EnvProvider string

// List of valid environment strategies
const (
	EnvProviderEcs        EnvProvider = "ecs"
	EnvProviderEcsFargate             = "ecs-fargate"
	EnvProviderEc2                    = "ec2"
	EnvProviderEks                    = "eks"
	EnvProviderEksFargate             = "eks-fargate"
)

// InstanceTenancy describes supported tenancy options for EC2
type InstanceTenancy string

// List of valid tenancies
const (
	InstanceTenancyDedicated = "dedicated"
	InstanceTenancyHost      = "host"
	InstanceTenancyDefault   = "default"
)

// ArtifactProvider describes supported artifact strategies
type ArtifactProvider string

// List of valid artifact providers
const (
	ArtifactProviderEcr ArtifactProvider = "ecr"
	ArtifactProviderS3                   = "s3"
)

// ServiceProtocol describes exposed ports for ECS service
type ServiceProtocol string

// List of supported service protocols
const (
	ServiceProtocolHTTP  = "HTTP"
	ServiceProtocolHTTPS = "HTTPS"
)

// NetworkMode describes the ecs docker network mode
type NetworkMode string

// List of supported network modes
const (
	NetworkModeNone   = "none"
	NetworkModeBridge = "bridge"
	NetworkModeAwsVpc = "awsvpc"
	NetworkModeHost   = "host"
)

// ComputeType describes the compute type of a codebuild project
type ComputeType string

// List of supported compute types
const (
	ComputeTypeSmall  = "BUILD_GENERAL1_SMALL"
	ComputeTypeMedium = "BUILD_GENERAL1_MEDIUM"
	ComputeTypeLarge  = "BUILD_GENERAL1_LARGE"
)

// EnvironmentType describes the codebuild project environment type
type EnvironmentType string

// List of supported environment types
const (
	EnvironmentTypeLinux   = "LINUX_CONTAINER"
	EnvironmentTypeWindows = "WINDOWS_CONTAINER"
)

// Container describes container details
type Container struct {
	Name     string
	Instance string
}

// Task describes task definition
type Task struct {
	Name           string
	Environment    string
	Service        string
	Status         string
	TaskDefinition string
	Cluster        string
	Command        []string
	Containers     []Container
}

// JSONOutput common json definition
type JSONOutput struct {
	Values [1]struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"values"`
}

// Int64Value returns the value of the int64 pointer passed in or
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

// StringValue returns the value of the string pointer passed in or
// "" if the pointer is nil.
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// StringRef returns the string pointer to the string passed in
func StringRef(v string) *string {
	return &v
}

// BoolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}

// TimeValue returns the value of the time.Time pointer passed in or
// time.Time{} if the pointer is nil.
func TimeValue(v *time.Time) time.Time {
	if v != nil {
		return *v
	}
	return time.Time{}
}

// CPUMemory represents valid cpu/memory structure
type CPUMemory struct {
	CPU    int
	Memory []int
}

// GB count of MB
var GB = 1024

// CPUMemorySupport represents valid ECS combinations
var CPUMemorySupport = []CPUMemory{
	{CPU: 256, Memory: []int{512, 1 * GB, 2 * GB}},
	{CPU: 512, Memory: []int{1 * GB, 2 * GB, 3 * GB, 4 * GB}},
	{CPU: 1024, Memory: []int{2 * GB, 3 * GB, 4 * GB, 5 * GB, 6 * GB, 7 * GB, 8 * GB}},
	{CPU: 2048, Memory: []int{4 * GB, 5 * GB, 6 * GB, 7 * GB, 8 * GB, 9 * GB, 10 * GB, 11 * GB, 12 * GB, 13 * GB, 14 * GB, 15 * GB, 16 * GB}},
	{CPU: 4096, Memory: []int{8 * GB, 9 * GB, 10 * GB, 11 * GB, 12 * GB, 13 * GB, 14 * GB, 15 * GB, 16 * GB, 17 * GB, 18 * GB, 19 * GB, 20 * GB,
		21 * GB, 22 * GB, 23 * GB, 24 * GB, 25 * GB, 26 * GB, 27 * GB, 28 * GB, 29 * GB, 30 * GB}},
}

// Warning that implements `error` but safe to ignore
type Warning struct {
	Message string
}

// Error the contract for error
func (w Warning) Error() string {
	return w.Message
}

// Warningf create a warning
func Warningf(format string, args ...interface{}) Warning {
	w := Warning{
		Message: fmt.Sprintf(format, args...),
	}
	return w
}
