package cli

import (
	"bytes"
	"flag"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"io/ioutil"
	"testing"
)

func TestNewEnvironmentsCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newEnvironmentsCommand(ctx)

	assert.NotNil(command)
	assert.Equal("environment", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("env", command.Aliases[0], "Aliases should match")
	assert.Equal("options for managing environments", command.Usage, "Usage should match")
	assert.Equal(5, len(command.Subcommands), "Subcommands len should match")

	args := []string{"environment", "help"}
	err := runCommand(command, args)
	assert.Nil(err)
}

func TestNewEnvironmentsUpsertCommand(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsUpsertCommand(ctx)

	assert.NotNil(command)
	assert.Equal("upsert", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("up", command.Aliases[0], "Aliases should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)

	args := []string{"upsert"}
	err := runCommand(command, args)
	assert.NotNil(err)
	assert.Equal(1, lastExitCode)

	args = []string{"upsert", "fooenv"}
	err = runCommand(command, args)
	assert.NotNil(err)
	assert.Equal(1, lastExitCode)
}

func TestNewEnvironmentsListCommand(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsListCommand(ctx)

	assert.NotNil(command)
	assert.Equal("list", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("ls", command.Aliases[0], "Aliases should match")
	assert.Equal("list environments", command.Usage, "Usage should match")
	assert.NotNil(command.Action)
}
func TestNewEnvironmentsShowCommand(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsShowCommand(ctx)

	assert.NotNil(command)
	assert.Equal("show", command.Name, "Name should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.Equal(1, len(command.Flags), "Flag len should match")
	assert.Equal("format, f", command.Flags[0].GetName(), "Flag should match")
	assert.NotNil(command.Action)
}
func TestNewEnvironmentsTerminateCommand(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	command := newEnvironmentsTerminateCommand(ctx)

	assert.NotNil(command)
	assert.Equal("terminate", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("term", command.Aliases[0], "Aliases should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}
func TestNewEnvironmentsLogsCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newEnvironmentsLogsCommand(ctx)

	assert.NotNil(command)
	assert.Equal("logs", command.Name, "Name should match")
	assert.Equal("<environment> [<filter>...]", command.ArgsUsage, "ArgsUsage should match")
	assert.Equal(2, len(command.Flags), "Flags length")
	assert.Equal("follow, f", command.Flags[0].GetName(), "Flags Name")
	assert.Equal("search-duration, t", command.Flags[1].GetName(), "Flags Name")
	assert.NotNil(command.Action)
}

func runCommand(command *cli.Command, args []string) error {
	app := cli.NewApp()
	app.Writer = ioutil.Discard
	set := flag.NewFlagSet("test", 0)
	set.Parse(args)
	appContext := cli.NewContext(app, set, nil)
	return command.Run(appContext)
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
