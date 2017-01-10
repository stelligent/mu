package cli

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stelligent/mu/common"
)

func TestNewPipelinesCommand(t *testing.T) {
	assert := assert.New(t)

	config := common.NewConfig()

	command := newPipelinesCommand(config)

	assert.NotNil(command)
	assert.Equal("pipeline", command.Name, "Name should match")
	assert.Equal("options for managing pipelines", command.Usage, "Usage should match")
	assert.Equal(2, len(command.Subcommands), "Subcommands len should match")
}
func TestNewPipelinesListCommand(t *testing.T) {
	assert := assert.New(t)

	config := common.NewConfig()

	command := newPipelinesListCommand(config)

	assert.NotNil(command)
	assert.Equal("list", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("ls", command.Aliases[0], "Aliases should match")
	assert.Equal("list pipelines", command.Usage, "Usage should match")
	assert.NotNil(command.Action)
}
func TestNewPipelinesShowCommand(t *testing.T) {
	assert := assert.New(t)

	config := common.NewConfig()

	command := newPipelinesShowCommand(config)

	assert.NotNil(command)
	assert.Equal("show", command.Name, "Name should match")
	assert.Equal(1, len(command.Flags), "Flag len should match")
	assert.Equal("service, s", command.Flags[0].GetName(), "Flag should match")
	assert.NotNil(command.Action)
}

