package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPipelinesCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newPipelinesCommand(ctx)

	assert.NotNil(command)
	assert.Equal("pipeline", command.Name, "Name should match")
	assert.Equal("options for managing pipelines", command.Usage, "Usage should match")
	assert.Equal(4, len(command.Subcommands), "Subcommands len should match")
}
func TestNewPipelinesListCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newPipelinesListCommand(ctx)

	assert.NotNil(command)
	assert.Equal("list", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("ls", command.Aliases[0], "Aliases should match")
	assert.Equal("list pipelines", command.Usage, "Usage should match")
	assert.NotNil(command.Action)
}
func TestNewPipelinesTerminateCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newPipelinesTerminateCommand(ctx)

	assert.NotNil(command)
	assert.Equal("terminate", command.Name, "Name should match")
	assert.Equal("[<service>]", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}
func TestNewPipelinesUpsertCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newPipelinesUpsertCommand(ctx)

	assert.NotNil(command)
	assert.Equal("upsert", command.Name, "Name should match")
	assert.Equal(1, len(command.Flags), "Flag len should match")
	assert.Equal("token, t", command.Flags[0].GetName(), "Flag should match")
	assert.NotNil(command.Action)
}
func TestNewPipelinesLogsCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newPipelinesLogsCommand(ctx)

	assert.NotNil(command)
	assert.Equal("logs", command.Name, "Name should match")
	assert.Equal("[<filter>...]", command.ArgsUsage, "ArgsUsage should match")
	assert.Equal(3, len(command.Flags), "Flags length")
	assert.Equal("service, s", command.Flags[0].GetName(), "Flags Name")
	assert.Equal("follow, f", command.Flags[1].GetName(), "Flags Name")
	assert.Equal("search-duration, t", command.Flags[2].GetName(), "Flags Name")
	assert.NotNil(command.Action)
}
