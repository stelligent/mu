package cli

import (
	"bytes"
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestNewEnvironmentsCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newEnvironmentsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(EnvCmd, command.Name, NameMessage)
	assertion.Equal(EnvAliasCount, len(command.Aliases), AliasLenMessage)
	assertion.Equal(EnvAlias, command.Aliases[SingleAliasIndex], AliasMessage)
	assertion.Equal(EnvUsage, command.Usage, UsageMessage)
	assertion.Equal(EnvSubCmdCount, len(command.Subcommands), SubCmdLenMessage)

	args := []string{EnvCmd, Help}
	err := runCommand(command, args)
	assertion.Nil(err)
}

func TestNewEnvironmentsUpsertCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsUpsertCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(UpsertCmd, command.Name, NameMessage)
	assertion.Equal(EnvAliasCount, len(command.Aliases), AliasLenMessage)
	assertion.Equal(UpsertAlias, command.Aliases[SingleAliasIndex], AliasMessage)
	assertion.Equal(EnvsArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.NotNil(command.Action)

	args := []string{UpsertCmd}
	err := runCommand(command, args)
	assertion.NotNil(err)
	assertion.Equal(FailExitCode, lastExitCode)

	lastExitCode = 0

	args = []string{UpsertCmd, TestEnv}
	err = runCommand(command, args)
	assertion.Nil(err)
	assertion.Equal(0, lastExitCode)
}

func TestNewEnvironmentsListCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsListCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(ListCmd, command.Name, NameMessage)
	assertion.Equal(EnvAliasCount, len(command.Aliases), AliasLenMessage)
	assertion.Equal(ListAlias, command.Aliases[SingleAliasIndex], AliasMessage)
	assertion.Equal(ListUsage, command.Usage, UsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsShowCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsShowCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(ShowCmd, command.Name, NameMessage)
	assertion.Equal(EnvArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(2, len(command.Flags), FlagLenMessage)
	assertion.Equal(FormatFlag, command.Flags[SvcShowFormatFlagIndex].GetName(), FlagMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsTerminateCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsTerminateCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(TerminateCmd, command.Name, NameMessage)
	assertion.Equal(EnvAliasCount, len(command.Aliases), AliasLenMessage)
	assertion.Equal(TerminateAlias, command.Aliases[SingleAliasIndex], AliasMessage)
	assertion.Equal(EnvsArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsLogsCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsLogsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(LogsCmd, command.Name, NameMessage)
	assertion.Equal(LogsArgs, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(EnvLogsFlagCount, len(command.Flags), FlagLenMessage)
	assertion.Equal(FollowFlag, command.Flags[EnvLogFollowFlagIndex].GetName(), FlagMessage)
	assertion.Equal(SearchDurationFlag, command.Flags[EnvLogDurationFlagIndex].GetName(), FlagMessage)
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
