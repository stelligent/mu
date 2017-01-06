package environments

import (
	"testing"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewTerminateCommand(t *testing.T) {
	assert := assert.New(t)

	config := &common.Config {}

	command := newTerminateCommand(config)

	assert.NotNil(command)
	assert.Equal("terminate", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("term", command.Aliases[0], "Aliases should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}

