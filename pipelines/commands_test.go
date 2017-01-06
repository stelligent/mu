package pipelines

import (
	"testing"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	assert := assert.New(t)

	config := &common.Config {}

	command := NewCommand(config)

	assert.NotNil(command)
	assert.Equal("pipeline", command.Name, "Name should match")
	assert.Equal("options for managing pipelines", command.Usage, "Usage should match")
	assert.Equal(2, len(command.Subcommands), "Subcommands len should match")
}