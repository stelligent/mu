package workflows

import (
	"io"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("workflows")

// Bold is the specifier for bold formatted text values
var Bold = color.New(color.Bold).SprintFunc()

// SvcPipelineTableHeader is the header array for the pipeline table
var SvcPipelineTableHeader = []string{SvcStageHeader, SvcActionHeader, SvcRevisionHeader, SvcStatusHeader, SvcLastUpdateHeader}

// SvcEnvironmentTableHeader is the header array for the environment table
var SvcEnvironmentTableHeader = []string{EnvironmentHeader, SvcRevisionHeader, SvcStatusHeader, SvcLastUpdateHeader}

// SvcTaskContainerHeader is the header for container task detail
var SvcTaskContainerHeader = []string{"Environment", "Container", "Task", "Instance"}

// PipeLineServiceHeader is the header for the pipeline service table
var PipeLineServiceHeader = []string{SvcServiceHeader, SvcStackHeader, SvcStatusHeader, SvcLastUpdateHeader}

// EnvironmentAMITableHeader is the header for the instance details
var EnvironmentAMITableHeader = []string{EC2Instance, TypeHeader, AMI, PrivateIP, AZ, ConnectedHeader, SvcStatusHeader, NumTasks, CPUAvail, MEMAvail}

// ServiceTableHeader is the header for the service table
var ServiceTableHeader = []string{SvcServiceHeader, SvcRevisionHeader, SvcStatusHeader, SvcLastUpdateHeader}

// EnvironmentShowHeader is the header for the environment table
var EnvironmentShowHeader = []string{EnvironmentHeader, SvcStackHeader, SvcStatusHeader, SvcLastUpdateHeader}

// Constants to prevent multiple updates when making changes.
const (
	Zero                   = 0
	FirstValueIndex        = 0
	PollDelay              = 5
	LineChar               = "-"
	NewLine                = "\n"
	NA                     = "N/A"
	UnknownValue           = "???"
	JSON                   = "json"
	SHELL                  = "shell"
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

// TagInterface used to conform tag structs
type TagInterface interface{}

// EnvironmentTags used to set defaults
type EnvironmentTags struct {
	Environment string `tag:"environment"`
	Type        string `tag:"type"`
	Provider    string `tag:"provider"`
	Revision    string `tag:"revision"`
	Repo        string `tag:"repo"`
}

// ServiceTags used to set defaults
type ServiceTags struct {
	Service     string `tag:"service"`
	Environment string `tag:"environment"`
	Type        string `tag:"type"`
	Provider    string `tag:"provider"`
	Revision    string `tag:"revision"`
	Repo        string `tag:"repo"`
}

// ScheduleTags used to set defaults
type ScheduleTags struct {
	Service     string `tag:"service"`
	Environment string `tag:"environment"`
	Type        string `tag:"type"`
}

// PipelineTags used to set defaults
type PipelineTags struct {
	Type     string `tag:"type"`
	Service  string `tag:"service"`
	Revision string `tag:"revision"`
	Repo     string `tag:"repo"`
}

// CatalogTags used to set defaults
type CatalogTags struct {
	Type string `tag:"type"`
}

// DatabaseTags used to set defaults
type DatabaseTags struct {
	Environment string `tag:"environment"`
	Type        string `tag:"type"`
	Service     string `tag:"service"`
	Revision    string `tag:"revision"`
	Repo        string `tag:"repo"`
}

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
func simplifyRepoURL(url string) string {
	slashIndex := strings.Index(url, "/")

	if slashIndex == -1 {
		return url
	}

	return url[slashIndex+1:]
}

func createTagMap(tagI TagInterface) map[string]string {
	rtn := map[string]string{}
	interfaceTags := map[string]string{}
	values := reflect.ValueOf(tagI).Elem()

	for i := 0; i < values.NumField(); i++ {
		interfaceTags[values.Type().Field(i).Tag.Get("tag")] = values.Field(i).String()
	}

	for key, value := range interfaceTags {
		rtn["mu:"+key] = value
	}

	return rtn
}
