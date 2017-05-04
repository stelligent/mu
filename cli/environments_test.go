package cli

import (
	"bytes"
	"errors"
	"flag"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"io/ioutil"
	"testing"
)

const (
	Test             = "test"
	TestEnv          = "fooenv"
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

func TestNewEnvironmentsCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newEnvironmentsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.EnvCmd, command.Name, NameMessage)
	assertion.Equal(1, len(command.Aliases), AliasLenMessage)
	assertion.Equal(common.EnvAlias, command.Aliases[0], AliasMessage)
	assertion.Equal(common.EnvUsage, command.Usage, UsageMessage)
	assertion.Equal(6, len(command.Subcommands), SubCmdLenMessage)

	args := []string{common.EnvCmd, Help}
	err := runCommand(command, args)
	assertion.Nil(err)
}

func TestNewEnvironmentsUpsertCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsUpsertCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.UpsertCmd, command.Name, NameMessage)
	assertion.Equal(1, len(command.Aliases), AliasLenMessage)
	assertion.Equal(common.UpsertAlias, command.Aliases[0], AliasMessage)
	assertion.Equal(common.EnvArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.NotNil(command.Action)

	args := []string{common.UpsertCmd}
	err := runCommand(command, args)
	assertion.NotNil(err)
	assertion.Equal(1, lastExitCode)

	args = []string{common.UpsertCmd, TestEnv}
	err = runCommand(command, args)
	assertion.NotNil(err)
	assertion.Equal(1, lastExitCode)
}

func testBaseEnvironmentExecute(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsExecuteCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.ExeCmd, command.Name, NameMessage)
	assertion.Equal(common.ExeArgs, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(common.ExeUsage, command.Usage, UsageMessage)
	assertion.NotNil(command.Action)
}

func getTestExecuteContext(args cli.Args) *cli.Context {
	app := cli.NewApp()
	app.Writer = ioutil.Discard
	set := flag.NewFlagSet(Test, 0)
	set.Parse(args)

	return cli.NewContext(app, set, nil)
}

func TestNewEnvironmentExecuteCommandNoEnv(t *testing.T) {
	assertion := assert.New(t)
	testBaseEnvironmentExecute(t)

	assertion.Equal(errors.New(common.NoEnvValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{})))
	assertion.Equal(errors.New(common.NoEnvValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{common.Spaces})))
}

func TestNewEnvironmentExecuteCommandNoCmd(t *testing.T) {
	assertion := assert.New(t)
	testBaseEnvironmentExecute(t)

	assertion.Equal(errors.New(common.NoCmdValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{TestEnv})))
	assertion.Equal(errors.New(common.EmptyCmdValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{TestEnv, common.Spaces})))
}

func TestNewEnvironmentExecuteCommand(t *testing.T) {
	assertion := assert.New(t)
	testBaseEnvironmentExecute(t)

	assertion.Nil(validateExecuteArguments(getTestExecuteContext(cli.Args{TestEnv, TestCmd})))
}

func TestNewEnvironmentsListCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsListCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.ListCmd, command.Name, NameMessage)
	assertion.Equal(1, len(command.Aliases), AliasLenMessage)
	assertion.Equal(common.ListAlias, command.Aliases[0], AliasMessage)
	assertion.Equal(common.ListUsage, command.Usage, UsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsShowCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsShowCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.ShowCmd, command.Name, NameMessage)
	assertion.Equal(common.EnvArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(1, len(command.Flags), FlagLenMessage)
	assertion.Equal(common.FormatFlag, command.Flags[0].GetName(), FlagMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsTerminateCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsTerminateCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.TerminateCmd, command.Name, NameMessage)
	assertion.Equal(1, len(command.Aliases), AliasLenMessage)
	assertion.Equal(common.TerminateAlias, command.Aliases[0], AliasMessage)
	assertion.Equal(common.EnvArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewEnvironmentsLogsCommand(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsLogsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.LogsCmd, command.Name, NameMessage)
	assertion.Equal(common.LogsArgs, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(2, len(command.Flags), FlagLenMessage)
	assertion.Equal(common.FollowFlag, command.Flags[0].GetName(), FlagMessage)
	assertion.Equal(common.SearchDurationFlag, command.Flags[1].GetName(), FlagMessage)
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
