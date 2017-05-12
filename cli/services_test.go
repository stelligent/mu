package cli

import (
	"errors"
	"flag"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"io/ioutil"
	"testing"
)

func TestNewServicesCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.SvcCmd, command.Name, common.NameMessage)
	assertion.Equal(common.SvcAliasCount, len(command.Aliases), common.AliasLenMessage)
	assertion.Equal(common.SvcAlias, command.Aliases[common.SingleAliasIndex], common.AliasMessage)
	assertion.Equal(common.SvcUsage, command.Usage, common.UsageMessage)
	assertion.Equal(common.SvcSubCmdCount, len(command.Subcommands), common.SubCmdLenMessage)
}

func TestNewServicesShowCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesShowCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.ShowCmd, command.Name, common.NameMessage)
	assertion.Equal(common.Zero, len(command.Flags), common.FlagLenMessage)
	assertion.Equal(common.SvcShowUsage, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewServicesPushCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesPushCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.PushCmd, command.Name, common.NameMessage)
	assertion.Equal(common.SvcAliasCount, len(command.Flags), common.FlagLenMessage)
	assertion.Equal(common.TagFlagName, command.Flags[common.SvcPushTagFlagIndex].GetName(), common.FlagMessage)
	assertion.NotNil(command.Action)
}

func TestNewServicesDeployCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesDeployCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.DeployCmd, command.Name, common.NameMessage)
	assertion.Equal(common.EnvArgUsage, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.Equal(common.SvcAliasCount, len(command.Flags), common.FlagLenMessage)
	assertion.Equal(common.TagFlagName, command.Flags[common.SvcDeployTagFlagIndex].GetName(), common.FlagMessage)
	assertion.NotNil(command.Action)
}

func TestNewUndeployCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesUndeployCommand(ctx)

	assertion.Equal(common.UndeployCmd, command.Name, common.NameMessage)
	assertion.Equal(common.SvcUndeployArgsUsage, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.Equal(common.Zero, len(command.Flags), common.FlagLenMessage)
	assertion.NotNil(command.Action)
}

func TestNewServicesLogsCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesLogsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.LogsCmd, command.Name, common.NameMessage)
	assertion.Equal(common.SvcLogArgUsage, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.Equal(common.SvcLogFlagCount, len(command.Flags), common.FlagLenMessage)
	assertion.Equal(common.ServiceFlag, command.Flags[common.SvcLogServiceFlagIndex].GetName(), common.FlagMessage)
	assertion.Equal(common.FollowFlag, command.Flags[common.SvcLogFollowFlagIndex].GetName(), common.FlagMessage)
	assertion.Equal(common.SearchDurationFlag, command.Flags[common.SvcLogDurationFlagIndex].GetName(), common.FlagMessage)
	assertion.NotNil(command.Action)
}

func TestExecuteTaskCreation(t *testing.T) {
	assertion := assert.New(t)
	args := []string{common.EnvCmd, common.Help}
	ctx := getTestExecuteContext(args)
	assertion.NotNil(ctx)
	task, err := newTask(ctx)
	assertion.NotNil(task)
	assertion.Nil(err)
}

func TestExecuteTaskCreationFail(t *testing.T) {
	assertion := assert.New(t)
	args := []string{}
	ctx := getTestExecuteContext(args)
	assertion.NotNil(ctx)
	task, err := newTask(ctx)
	assertion.Nil(task)
	assertion.NotNil(err)
}

func TestNewServiceExecuteCommandNoEnv(t *testing.T) {
	assertion := assert.New(t)
	testBaseServiceExecute(t)

	assertion.Equal(errors.New(common.NoEnvValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{})))
	assertion.Equal(errors.New(common.NoEnvValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{common.Spaces})))
}

func TestNewServiceExecuteCommandNoCmd(t *testing.T) {
	assertion := assert.New(t)
	testBaseServiceExecute(t)

	assertion.Equal(errors.New(common.NoCmdValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{common.TestEnv})))
	assertion.Equal(errors.New(common.EmptyCmdValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{common.TestEnv, common.Spaces})))
}

func TestNewServiceExecuteCommand(t *testing.T) {
	assertion := assert.New(t)
	testBaseServiceExecute(t)

	assertion.Nil(validateExecuteArguments(getTestExecuteContext(cli.Args{common.TestEnv, common.TestSvc, common.TestCmd})))
}

func testBaseServiceExecute(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newServicesExecuteCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(common.ExeCmd, command.Name, common.NameMessage)
	assertion.Equal(common.ExeArgs, command.ArgsUsage, common.ArgsUsageMessage)
	assertion.Equal(common.ExeUsage, command.Usage, common.UsageMessage)
	assertion.NotNil(command.Action)
}

func getTestExecuteContext(args cli.Args) *cli.Context {
	app := cli.NewApp()
	app.Writer = ioutil.Discard
	set := flag.NewFlagSet(common.Test, common.Zero)
	set.Parse(args)

	return cli.NewContext(app, set, nil)
}
