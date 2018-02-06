package aws

import "github.com/op/go-logging"

var log = logging.MustGetLogger("aws")

// Constants to prevent multiple updates when making changes.
const (
	Zero                        = 0
	Empty                       = ""
	ForwardSlash                = "/"
	ECSServiceNameParameterKey  = "ServiceName"
	ListServices                = "ListServices"
	ListTasks                   = "ListTasks"
	DescribeTasks               = "DescribeTasks"
	DescribeContainerInstances  = "DescribeContainerInstances"
	ECSTaskDefinitionOutputKey  = "MicroserviceTaskDefinitionArn"
	ECSClusterOutputKey         = "EcsCluster"
	SvcCmdStackLog              = "Getting stack '%s'..."
	EcsConnectionLog            = "Connecting to ECS service"
	ExecuteCommandStartLog      = "Executing command '[%s]' on environment '%s' for service '%s'\n"
	ExecuteCommandFinishLog     = "Command execution complete\n"
	ExecuteECSInputParameterLog = "Environment: %s, Service: %s, Cluster: %s, Task: %s"
	ExecuteECSInputContentsLog  = "ECS Input Contents: %s\n"
	ExecuteECSResultContentsLog = "ECS Result Contents: %s, %s\n"
	SvcGetTaskInfoLog           = "Getting task info for task: %s"
	SvcTaskDetailLog            = "Task Detail: %s"
	SvcListTasksLog             = "Listing tasks for Environment: %s, Cluster: %s, Service: %s"
	TaskARNSeparator            = ForwardSlash
)

// Constants used during testing
const (
	TestEnv      = "fooenv"
	TestSvc      = "foosvc"
	TestCmd      = "foocmd"
	TestTaskARN  = "ARN/TEST"
	GetStackName = "GetStack"
	RunTaskName  = "RunTask"
)
