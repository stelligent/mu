package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDatabasesCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newDatabasesCommand(ctx)

	assert.NotNil(command)
	assert.Equal("database", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("db", command.Aliases[0], "Aliases should match")
	assert.Equal("options for managing databases", command.Usage, "Usage should match")
	assert.Equal(3, len(command.Subcommands), "Subcommands len should match")
}

func TestNewDatabasesUpsertCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newDatabaseUpsertCommand(ctx)

	assert.NotNil(command)
	assert.Equal("upsert", command.Name, "Name should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}

func TestNewDatabaseTerminateCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newDatabaseTerminateCommand(ctx)

	assert.Equal("terminate", command.Name, "Name should match")
	assert.Equal("<environment> [<service>]", command.ArgsUsage, "ArgsUsage should match")
	assert.Equal(0, len(command.Flags), "Flags length")
	assert.NotNil(command.Action)
}

func TestNewDatabaseListCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newDatabaseListCommand(ctx)

	assert.NotNil(command)
	assert.Equal("list", command.Name, "Name should match")
	assert.NotNil(command.Action)
}
