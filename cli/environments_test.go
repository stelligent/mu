package cli

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stelligent/mu/common"
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
	assert.Equal(4, len(command.Subcommands), "Subcommands len should match")
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
