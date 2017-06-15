package common

import (
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"io"
	"time"
)

// Bold is the specifier for bold formatted text values
var Bold = color.New(color.Bold).SprintFunc()

// SvcPipelineTableHeader is the header array for the pipeline table
var SvcPipelineTableHeader = []string{SvcStageHeader, SvcActionHeader, SvcRevisionHeader, SvcStatusHeader, SvcLastUpdateHeader}

// SvcEnvironmentTableHeader is the header array for the environment table
var SvcEnvironmentTableHeader = []string{EnvironmentHeader, SvcStackHeader, SvcImageHeader, SvcStatusHeader, SvcLastUpdateHeader, SvcMuVersionHeader}

// SvcTaskContainerHeader is the header for container task detail
var SvcTaskContainerHeader = []string{"Task", "Container", "Instance"}

// PipeLineServiceHeader is the header for the pipeline service table
var PipeLineServiceHeader = []string{SvcServiceHeader, SvcStackHeader, SvcStatusHeader, SvcLastUpdateHeader, SvcMuVersionHeader}

// EnvironmentAMITableHeader is the header for the instance details
var EnvironmentAMITableHeader = []string{EC2Instance, TypeHeader, AMI, PrivateIP, AZ, ConnectedHeader, SvcStatusHeader, NumTasks, CPUAvail, MEMAvail}

// ServiceTableHeader is the header for the service table
var ServiceTableHeader = []string{SvcServiceHeader, SvcImageHeader, SvcStatusHeader, SvcLastUpdateHeader, SvcMuVersionHeader}

// EnvironmentShowHeader is the header for the environment table
var EnvironmentShowHeader = []string{EnvironmentHeader, SvcStackHeader, SvcStatusHeader, SvcLastUpdateHeader, SvcMuVersionHeader}

// Constants for available command names and options
const (
	EnvSubCmdCount          = 5
	FirstValueIndex         = 0
	SingleAliasIndex        = 0
	SvcSubCmdCount          = 6
	SvcShowFormatFlagIndex  = 0
	SvcLogFlagCount         = 3
	EnvLogFollowFlagIndex   = 0
	EnvLogDurationFlagIndex = 1
	SvcLogServiceFlagIndex  = 0
	SvcLogFollowFlagIndex   = 1
	SvcLogDurationFlagIndex = 2
	ShowFlagCount           = 1
	ExeArgsCmdIndex         = 1
	EnvLogsFlagCount        = 2
	SvcPushTagFlagIndex     = 0
	SvcDeployTagFlagIndex   = 0
	SvcUndeploySvcFlagIndex = 1
	TaskGUIDIndex           = 1
	DefaultLogDurationValue = 1 * time.Minute
	SvcCmd                  = "service"
	SvcAlias                = "svc"
	SvcUsage                = "options for managing services"
	SvcShowUsage            = "[<service>]"
	SvcLogUsage             = "show service logs"
	SvcLogArgUsage          = "<environment> [<filter>...]"
	SvcLogServiceFlagUsage  = "service name to view logs for"
	SvcExeServiceFlagUsage  = "service name for command"
	SvcExeTaskFlagUsage     = "task definition arn"
	SvcExeClusterFlagUsage  = "cluster name or full arn"
	SvcPushTagFlagUsage     = "tag to push"
	SvcDeployTagFlagUsage   = "tag to deploy"
	TagFlagName             = "tag, t"
	EnvCmd                  = "environment"
	EnvAlias                = "env"
	EnvUsage                = "options for managing environments"
	EnvArgUsage             = "<environment>"
	Tag                     = "tag"
	UpsertCmd               = "upsert"
	UpsertAlias             = "up"
	UpsertUsage             = "create/update an environment"
	ListCmd                 = "list"
	TerminateCmd            = "terminate"
	TerminateAlias          = "term"
	TerminateUsage          = "terminate an environment"
	ListAlias               = "ls"
	ListUsage               = "list environments"
	ShowCmd                 = "show"
	ShowCmdUsage            = "show environment details"
	ExeCmd                  = "exec"
	ExeUsage                = "execute a command in environment"
	ExeArgs                 = "<environment> <command>"
	LogsCmd                 = "logs"
	LogsArgs                = "<environment> [<filter>...]"
	LogsUsage               = "show environment logs"
	Format                  = "format"
	FormatFlag              = "format, f"
	FormatFlagUsage         = "output format, either 'json' or 'cli' (default: cli)"
	FormatFlagDefault       = "cli"
	Follow                  = "follow"
	FollowFlag              = "follow, f"
	ServiceFlag             = "service, s"
	TaskFlagName            = "task"
	TaskFlagVisible         = true
	TaskFlag                = "task, t"
	ClusterFlagName         = "cluster"
	ClusterFlag             = "cluster, c"
	ClusterFlagVisible      = true
	FollowUsage             = "follow logs for latest changes"
	SearchDuration          = "search-duration"
	SearchDurationUsage     = "duration to go into the past for searching (e.g. 5m for 5 minutes)"
	SearchDurationFlag      = "search-duration, t"
	PushCmd                 = "push"
	SvcPushCmdUsage         = "push service to repository"
	DeployCmd               = "deploy"
	SvcDeployCmdUsage       = "deploy service to environment"
	UndeployCmd             = "undeploy"
	SvcUndeployCmdUsage     = "undeploy service from environment"
	SvcUndeployArgsUsage    = "<environment> [<service>]"
)

// Constants to prevent multiple updates when making changes.
const (
	Zero                        = 0
	ECSRunTaskDefaultCount      = 1
	Empty                       = ""
	Space                       = " "
	Spaces                      = "   "
	LineChar                    = "-"
	ForwardSlash                = "/"
	NewLine                     = "\n"
	NA                          = "N/A"
	UnknownValue                = "???"
	JSON                        = "json"
	HomeIPAddress               = "127.0.0.1"
	DefaultVersion              = "0.0.0-local"
	LastUpdateTime              = "2006-01-02 15:04:05"
	CPU                         = "CPU"
	MEMORY                      = "MEMORY"
	AMI                         = "AMI"
	PrivateIP                   = "IP Address"
	AZ                          = "AZ"
	BoolStringFormat            = "%v"
	IntStringFormat             = "%d"
	HeaderValueFormat           = "%s:\t%s\n"
	SvcPipelineFormat           = HeaderValueFormat
	HeadNewlineHeader           = "%s:\n"
	SvcDeploymentsFormat        = HeadNewlineHeader
	SvcContainersFormat         = "\n%s for %s:\n"
	KeyValueFormat              = "%s %s"
	StackFormat                 = "%s:\t%s (%s)\n"
	UnmanagedStackFormat        = "%s:\tunmanaged\n"
	BaseURLKey                  = "BASE_URL"
	BaseURLValueKey             = "BaseUrl"
	SvcPipelineURLLabel         = "Pipeline URL"
	SvcDeploymentsLabel         = "Deployments"
	SvcContainersLabel          = "Containers"
	BaseURLHeader               = "Base URL"
	SvcCodePipelineURLKey       = "CodePipelineUrl"
	SvcVersionKey               = "version"
	SvcCodePipelineNameKey      = "PipelineName"
	ECSClusterKey               = "EcsCluster"
	EC2Instance                 = "EC2 Instance"
	VPCStack                    = "VPC Stack"
	ContainerInstances          = "Container Instances"
	BastionHost                 = "Bastion Host"
	BastionHostKey              = "BastionHost"
	ClusterStack                = "Cluster Stack"
	TypeHeader                  = "Type"
	ConnectedHeader             = "Connected"
	CPUAvail                    = "CPU Avail"
	MEMAvail                    = "Mem Avail"
	NumTasks                    = "# Tasks"
	SvcImageURLKey              = "ImageUrl"
	SvcStageHeader              = "Stage"
	SvcServiceHeader            = "Service"
	ServicesHeader              = "Services"
	SvcActionHeader             = "Action"
	SvcStatusHeader             = "Status"
	SvcRevisionHeader           = "Revision"
	SvcMuVersionHeader          = "Mu Version"
	SvcImageHeader              = "Image"
	EnvironmentHeader           = "Environment"
	SvcStackHeader              = "Stack"
	SvcLastUpdateHeader         = "Last Update"
	ECSServiceNameParameterKey  = "ServiceName"
	ListServices                = "ListServices"
	DescribeInstances           = "DescribeInstances"
	ListTasks                   = "ListTasks"
	DescribeTasks               = "DescribeTasks"
	DescribeContainerInstances  = "DescribeContainerInstances"
	ECSTaskDefinitionOutputKey  = "MicroserviceTaskDefinition"
	ECSClusterOutputKey         = "EcsCluster"
	NoEnvValidation             = "environment must be provided"
	NoCmdValidation             = "command must be provided"
	EmptyCmdValidation          = "command must not be an empty string"
	SvcCmdTaskExecutingLog      = "Creating service executor...\n"
	SvcCmdTaskResultLog         = "Service executor complete with result:\n%s\n"
	SvcCmdStackLog              = "Getting stack '%s'..."
	SvcCmdTaskErrorLog          = "The following error has occurred executing the command:  '%v'"
	EcsConnectionLog            = "Connecting to ECS service"
	ExecuteCommandStartLog      = "Executing command '[%s]' on environment '%s' for service '%s'\n"
	ExecuteCommandFinishLog     = "Command execution complete\n"
	ExecuteECSInputParameterLog = "Environment: %s, Service: %s, Cluster: %s, Task: %s"
	ExecuteECSInputContentsLog  = "ECS Input Contents: %s\n"
	ExecuteECSResultContentsLog = "ECS Result Contents: %s, %s\n"
	SvcGetTaskInfoLog           = "Getting task info for task: %s"
	SvcTaskDetailLog            = "Task Detail: %s"
	SvcInstancePrivateIPLog     = "Instance Private IP for Instance ID %s: %s"
	SvcListTasksLog             = "Listing tasks for Environment: %s, Cluster: %s, Service: %s"
	ECSAvailabilityZoneKey      = "ecs.availability-zone"
	ECSInstanceTypeKey          = "ecs.instance-type"
	ECSAMIKey                   = "ecs.ami-id"
	TaskARNSeparator            = ForwardSlash
)

// Constants used during testing
const (
	EnvAliasCount    = 1
	SvcAliasCount    = 1
	FailExitCode     = 1
	Test             = "test"
	TestEnv          = "fooenv"
	TestSvc          = "foosvc"
	TestCmd          = "foocmd"
	TestTaskARN      = "ARN/TEST"
	Help             = "help"
	GetStackName     = "GetStack"
	RunTaskName      = "RunTask"
	NameMessage      = "Name should match"
	UsageMessage     = "Usage should match"
	AliasLenMessage  = "Aliases len should match"
	AliasMessage     = "Aliases should match"
	ArgsUsageMessage = "ArgsUsage should match"
	SubCmdLenMessage = "Subcommands len should match"
	FlagLenMessage   = "Flag len should match"
	FlagMessage      = "Flag should match"
)

// CreateTableSection creates the standard output table used
func CreateTableSection(writer io.Writer, header []string) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetBorder(true)
	table.SetAutoWrapText(false)
	return table
}
