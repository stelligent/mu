package workflows

import (
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/op/go-logging"
	"io"
)

var log = logging.MustGetLogger("workflows")

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

// Constants to prevent multiple updates when making changes.
const (
	Zero                   = 0
	FirstValueIndex        = 0
	LineChar               = "-"
	NewLine                = "\n"
	NA                     = "N/A"
	UnknownValue           = "???"
	JSON                   = "json"
	LastUpdateTime         = "2006-01-02 15:04:05"
	CPU                    = "CPU"
	MEMORY                 = "MEMORY"
	AMI                    = "AMI"
	PrivateIP              = "IP Address"
	AZ                     = "AZ"
	BoolStringFormat       = "%v"
	IntStringFormat        = "%d"
	HeaderValueFormat      = "%s:\t%s\n"
	SvcPipelineFormat      = HeaderValueFormat
	HeadNewlineHeader      = "%s:\n"
	SvcDeploymentsFormat   = HeadNewlineHeader
	SvcContainersFormat    = "\n%s for %s:\n"
	KeyValueFormat         = "%s %s"
	StackFormat            = "%s:\t%s (%s)\n"
	UnmanagedStackFormat   = "%s:\tunmanaged\n"
	BaseURLKey             = "BASE_URL"
	BaseURLValueKey        = "BaseUrl"
	SvcPipelineURLLabel    = "Pipeline URL"
	SvcDeploymentsLabel    = "Deployments"
	SvcContainersLabel     = "Containers"
	BaseURLHeader          = "Base URL"
	EnvTagKey              = "environment"
	SvcTagKey              = "service"
	SvcCodePipelineURLKey  = "CodePipelineUrl"
	SvcVersionKey          = "version"
	SvcCodePipelineNameKey = "PipelineName"
	ECSClusterKey          = "EcsCluster"
	EC2Instance            = "EC2 Instance"
	VPCStack               = "VPC Stack"
	ContainerInstances     = "Container Instances"
	BastionHost            = "Bastion Host"
	BastionHostKey         = "BastionHost"
	ClusterStack           = "Cluster Stack"
	TypeHeader             = "Type"
	ConnectedHeader        = "Connected"
	CPUAvail               = "CPU Avail"
	MEMAvail               = "Mem Avail"
	NumTasks               = "# Tasks"
	SvcImageURLKey         = "ImageUrl"
	SvcStageHeader         = "Stage"
	SvcServiceHeader       = "Service"
	ServicesHeader         = "Services"
	SvcActionHeader        = "Action"
	SvcStatusHeader        = "Status"
	SvcRevisionHeader      = "Revision"
	SvcMuVersionHeader     = "Mu Version"
	SvcImageHeader         = "Image"
	EnvironmentHeader      = "Environment"
	SvcStackHeader         = "Stack"
	SvcLastUpdateHeader    = "Last Update"
	SvcCmdTaskExecutingLog = "Creating service executor...\n"
	SvcCmdTaskResultLog    = "Service executor complete with result:\n%s\n"
	SvcCmdTaskErrorLog     = "The following error has occurred executing the command:  '%v'"
	ECSAvailabilityZoneKey = "ecs.availability-zone"
	ECSInstanceTypeKey     = "ecs.instance-type"
	ECSAMIKey              = "ecs.ami-id"
)

// Constants used during testing
const (
	TestEnv = "fooenv"
	TestSvc = "foosvc"
	TestCmd = "foocmd"
)

// CreateTableSection creates the standard output table used
func CreateTableSection(writer io.Writer, header []string) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(header)
	table.SetBorder(true)
	table.SetAutoWrapText(false)
	return table
}
