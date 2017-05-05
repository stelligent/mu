package cli

import (
	"bytes"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"testing"
)

func TestNewEnvironmentsCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newEnvironmentsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.EnvCmd, command.Name, common.NameMessage)
	assertion.Equal(common.EnvAliasCount, len(command.Aliases), common.AliasLenMessage)
	assertion.Equal(common.EnvAlias, command.Aliases[common.SingleAliasIndex], common.AliasMessage)
	assertion.Equal(common.EnvUsage, command.Usage, common.UsageMessage)
	assertion.Equal(common.EnvSubCmdCount, len(command.Subcommands), common.SubCmdLenMessage)

	args := []string{common.EnvCmd, common.Help}
	err := runCommand(command, args)
	assertion.Nil(err)
}

func TestNewEnvironmentsUpsertCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsUpsertCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.UpsertCmd, command.Name, common.NameMessage)
	assertion.Equal(common.EnvAliasCount, len(command.Aliases), common.AliasLenMessage)
	assertion.Equal(common.UpsertAlias, command.Aliases[common.SingleAliasIndex], common.AliasMessage)
	assertion.Equal(common.EnvArgUsage, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.NotNil(command.Action)

	args := []string{common.UpsertCmd}
	err := runCommand(command, args)
	assertion.NotNil(err)
	assertion.Equal(common.FailExitCode, lastExitCode)

	args = []string{common.UpsertCmd, common.TestEnv}
	err = runCommand(command, args)
	assertion.NotNil(err)
	assertion.Equal(common.FailExitCode, lastExitCode)
}

func TestNewEnvironmentsListCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsListCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.ListCmd, command.Name, common.NameMessage)
	assertion.Equal(common.EnvAliasCount, len(command.Aliases), common.AliasLenMessage)
	assertion.Equal(common.ListAlias, command.Aliases[common.SingleAliasIndex], common.AliasMessage)
	assertion.Equal(common.ListUsage, command.Usage, common.UsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsShowCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsShowCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.ShowCmd, command.Name, common.NameMessage)
	assertion.Equal(common.EnvArgUsage, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.Equal(common.ShowFlagCount, len(command.Flags), common.FlagLenMessage)
	assertion.Equal(common.FormatFlag, command.Flags[common.SvcShowFormatFlagIndex].GetName(), common.FlagMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsTerminateCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsTerminateCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.TerminateCmd, command.Name, common.NameMessage)
	assertion.Equal(common.EnvAliasCount, len(command.Aliases), common.AliasLenMessage)
	assertion.Equal(common.TerminateAlias, command.Aliases[common.SingleAliasIndex], common.AliasMessage)
	assertion.Equal(common.EnvArgUsage, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsLogsCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsLogsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.LogsCmd, command.Name, common.NameMessage)
	assertion.Equal(common.LogsArgs, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.Equal(common.EnvLogsFlagCount, len(command.Flags), common.FlagLenMessage)
	assertion.Equal(common.FollowFlag, command.Flags[common.EnvLogFollowFlagIndex].GetName(), common.FlagMessage)
	assertion.Equal(common.SearchDurationFlag, command.Flags[common.EnvLogDurationFlagIndex].GetName(), common.FlagMessage)
	assertion.NotNil(command.Action)
}

func runCommand(command *cli.Command, args []string) error {
	return command.Run(getTestExecuteContext(args))
}

var (
	lastExitCode = 0
	fakeOsExiter = func(rc int) {
		lastExitCode = rc
	}
	fakeErrWriter = &bytes.Buffer{}
)

func init() {
	cli.OsExiter = fakeOsExiter
	cli.ErrWriter = fakeErrWriter
}

type mockedCloudFormation struct {
	cloudformationiface.CloudFormationAPI
}
