package cli

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/resources"
)

func TestNewEnvironmentsCommand(t *testing.T) {
	assert := assert.New(t)

	config := common.NewConfig()

	command := newEnvironmentsCommand(config)

	assert.NotNil(command)
	assert.Equal("environment", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("env", command.Aliases[0], "Aliases should match")
	assert.Equal("options for managing environments", command.Usage, "Usage should match")
	assert.Equal(4, len(command.Subcommands), "Subcommands len should match")
}

func TestNewEnvironmentsUpsertCommand(t *testing.T) {
	assert := assert.New(t)
	config := common.NewConfig()
	envMgr := resources.NewEnvironmentManager(config)
	command := newEnvironmentsUpsertCommand(envMgr)

	assert.NotNil(command)
	assert.Equal("upsert", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("up", command.Aliases[0], "Aliases should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}

func TestNewEnvironmentsListCommand(t *testing.T) {
	assert := assert.New(t)
	config := common.NewConfig()
	envMgr := resources.NewEnvironmentManager(config)
	command := newEnvironmentsListCommand(envMgr)

	assert.NotNil(command)
	assert.Equal("list", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("ls", command.Aliases[0], "Aliases should match")
	assert.Equal("list environments", command.Usage, "Usage should match")
	assert.NotNil(command.Action)
}
func TestNewEnvironmentsShowCommand(t *testing.T) {
	assert := assert.New(t)
	config := common.NewConfig()
	envMgr := resources.NewEnvironmentManager(config)
	command := newEnvironmentsShowCommand(envMgr)

	assert.NotNil(command)
	assert.Equal("show", command.Name, "Name should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}
func TestNewEnvironmentsTerminateCommand(t *testing.T) {
	assert := assert.New(t)
	config := common.NewConfig()
	envMgr := resources.NewEnvironmentManager(config)
	command := newEnvironmentsTerminateCommand(envMgr)

	assert.NotNil(command)
	assert.Equal("terminate", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("term", command.Aliases[0], "Aliases should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}
