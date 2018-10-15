package cli

import (
	"errors"
	"flag"
	"io/ioutil"
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestNewServicesCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(SvcCmd, command.Name, NameMessage)
	assertion.Equal(SvcAliasCount, len(command.Aliases), AliasLenMessage)
	assertion.Equal(SvcAlias, command.Aliases[SingleAliasIndex], AliasMessage)
	assertion.Equal(SvcUsage, command.Usage, UsageMessage)
	assertion.Equal(SvcSubCmdCount, len(command.Subcommands), SubCmdLenMessage)
}

func TestNewServicesShowCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesShowCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(ShowCmd, command.Name, NameMessage)
	assertion.Equal(2, len(command.Flags), FlagLenMessage)
	assertion.Equal(SvcShowUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.NotNil(command.Action)
}

func TestNewServicesPushCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesPushCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(PushCmd, command.Name, NameMessage)
	assertion.Equal(SvcFlagsCount, len(command.Flags), FlagLenMessage)
	assertion.Equal(TagFlagName, command.Flags[SvcPushTagFlagIndex].GetName(), FlagMessage)
	assertion.NotNil(command.Action)
}

func TestNewServicesDeployCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesDeployCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(DeployCmd, command.Name, NameMessage)
	assertion.Equal(EnvArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(SvcAliasCount, len(command.Flags), FlagLenMessage)
	assertion.Equal(TagFlagName, command.Flags[SvcDeployTagFlagIndex].GetName(), FlagMessage)
	assertion.NotNil(command.Action)
}

func TestNewUndeployCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesUndeployCommand(ctx)

	assertion.Equal(UndeployCmd, command.Name, NameMessage)
	assertion.Equal(SvcUndeployArgsUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(Zero, len(command.Flags), FlagLenMessage)
	assertion.NotNil(command.Action)
}

func TestNewServiceRestartCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesRestartCommand(ctx)

	assertion.Equal(RestartCmd, command.Name, NameMessage)
	assertion.Equal(EnvArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(2, len(command.Flags), FlagLenMessage)
	assertion.NotNil(command.Action)
}

func TestNewServicesLogsCommand(t *testing.T) {
	assertion := assert.New(t)

	ctx := common.NewContext()

	command := newServicesLogsCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(LogsCmd, command.Name, NameMessage)
	assertion.Equal(SvcLogArgUsage, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(SvcLogFlagCount, len(command.Flags), FlagLenMessage)
	assertion.Equal(ServiceFlag, command.Flags[SvcLogServiceFlagIndex].GetName(), FlagMessage)
	assertion.Equal(FollowFlag, command.Flags[SvcLogFollowFlagIndex].GetName(), FlagMessage)
	assertion.Equal(SearchDurationFlag, command.Flags[SvcLogDurationFlagIndex].GetName(), FlagMessage)
	assertion.NotNil(command.Action)
}

func TestExecuteTaskCreation(t *testing.T) {
	assertion := assert.New(t)
	args := []string{EnvCmd, Help}
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

	assertion.Equal(errors.New(NoEnvValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{})))
	assertion.Equal(errors.New(NoEnvValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{Spaces})))
}

func TestNewServiceExecuteCommandNoCmd(t *testing.T) {
	assertion := assert.New(t)
	testBaseServiceExecute(t)

	assertion.Equal(errors.New(NoCmdValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{TestEnv})))
	assertion.Equal(errors.New(EmptyCmdValidation), validateExecuteArguments(getTestExecuteContext(cli.Args{TestEnv, Spaces})))
}

func TestNewServiceExecuteCommand(t *testing.T) {
	assertion := assert.New(t)
	testBaseServiceExecute(t)

	assertion.Nil(validateExecuteArguments(getTestExecuteContext(cli.Args{TestEnv, TestSvc, TestCmd})))
}

func testBaseServiceExecute(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	command := newServicesExecuteCommand(ctx)

	assertion.NotNil(command)
	assertion.Equal(ExeCmd, command.Name, NameMessage)
	assertion.Equal(ExeArgs, command.ArgsUsage, ArgsUsageMessage)
	assertion.Equal(ExeUsage, command.Usage, UsageMessage)
	assertion.NotNil(command.Action)
}

func getTestExecuteContext(args cli.Args) *cli.Context {
	app := cli.NewApp()
	app.Writer = ioutil.Discard
	set := flag.NewFlagSet(Test, Zero)
	set.Parse(args)

	return cli.NewContext(app, set, nil)
}
