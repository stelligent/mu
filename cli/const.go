package cli

import (
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("cli")

// Constants for available command names and options
const (
	EnvSubCmdCount           = 5
	SingleAliasIndex         = 0
	SvcSubCmdCount           = 6
	SvcShowFormatFlagIndex   = 0
	SvcLogFlagCount          = 3
	EnvLogFollowFlagIndex    = 0
	EnvLogDurationFlagIndex  = 1
	SvcLogServiceFlagIndex   = 0
	SvcLogFollowFlagIndex    = 1
	SvcLogDurationFlagIndex  = 2
	ShowFlagCount            = 1
	ExeArgsCmdIndex          = 1
	EnvLogsFlagCount         = 2
	SvcPushTagFlagIndex      = 0
	SvcDeployTagFlagIndex    = 0
	SvcUndeploySvcFlagIndex  = 1
	DefaultLogDurationValue  = 1 * time.Minute
	SvcCmd                   = "service"
	SvcAlias                 = "svc"
	SvcUsage                 = "options for managing services"
	SvcShowUsage             = "[<service>]"
	SvcLogUsage              = "show service logs"
	SvcLogArgUsage           = "<environment> [<filter>...]"
	SvcLogServiceFlagUsage   = "service name to view logs for"
	SvcExeServiceFlagUsage   = "service name for command"
	SvcExeTaskFlagUsage      = "task definition arn"
	SvcExeClusterFlagUsage   = "cluster name or full arn"
	SvcPushTagFlagUsage      = "tag to push"
	SvcPushProviderFlagUsage = "provider to push to"
	SvcDeployTagFlagUsage    = "tag to deploy"
	TagFlagName              = "tag, t"
	ProviderFlagName         = "provider, p"
	EnvCmd                   = "environment"
	EnvAlias                 = "env"
	EnvUsage                 = "options for managing environments"
	EnvArgUsage              = "<environment>"
	Tag                      = "tag"
	Provider                 = "provider"
	UpsertCmd                = "upsert"
	UpsertAlias              = "up"
	UpsertUsage              = "create/update an environment"
	ListCmd                  = "list"
	TerminateCmd             = "terminate"
	TerminateAlias           = "term"
	TerminateUsage           = "terminate an environment"
	ListAlias                = "ls"
	ListUsage                = "list environments"
	ShowCmd                  = "show"
	ShowCmdUsage             = "show environment details"
	ExeCmd                   = "exec"
	ExeUsage                 = "execute a command in environment"
	ExeArgs                  = "<environment> <command>"
	LogsCmd                  = "logs"
	LogsArgs                 = "<environment> [<filter>...]"
	LogsUsage                = "show environment logs"
	Format                   = "format"
	FormatFlag               = "format, f"
	FormatFlagUsage          = "output format, either 'json' or 'cli' (default: cli)"
	FormatFlagDefault        = "cli"
	Follow                   = "follow"
	FollowFlag               = "follow, f"
	ServiceFlag              = "service, s"
	TaskFlagName             = "task"
	TaskFlagVisible          = true
	TaskFlag                 = "task, t"
	ClusterFlagName          = "cluster"
	ClusterFlag              = "cluster, c"
	ClusterFlagVisible       = true
	FollowUsage              = "follow logs for latest changes"
	SearchDuration           = "search-duration"
	SearchDurationUsage      = "duration to go into the past for searching (e.g. 5m for 5 minutes)"
	SearchDurationFlag       = "search-duration, t"
	PushCmd                  = "push"
	SvcPushCmdUsage          = "push service to repository"
	DeployCmd                = "deploy"
	SvcDeployCmdUsage        = "deploy service to environment"
	UndeployCmd              = "undeploy"
	SvcUndeployCmdUsage      = "undeploy service from environment"
	SvcUndeployArgsUsage     = "<environment> [<service>]"
)

// Constants to prevent multiple updates when making changes.
const (
	Zero               = 0
	Space              = " "
	Spaces             = "   "
	NoEnvValidation    = "environment must be provided"
	NoCmdValidation    = "command must be provided"
	EmptyCmdValidation = "command must not be an empty string"
)

// Constants used during testing
const (
	EnvAliasCount    = 1
	SvcAliasCount    = 1
	SvcFlagsCount    = 2
	FailExitCode     = 1
	Test             = "test"
	TestEnv          = "fooenv"
	TestSvc          = "foosvc"
	TestCmd          = "foocmd"
	Help             = "help"
	NameMessage      = "Name should match"
	UsageMessage     = "Usage should match"
	AliasLenMessage  = "Aliases len should match"
	AliasMessage     = "Aliases should match"
	ArgsUsageMessage = "ArgsUsage should match"
	SubCmdLenMessage = "Subcommands len should match"
	FlagLenMessage   = "Flag len should match"
	FlagMessage      = "Flag should match"
)
