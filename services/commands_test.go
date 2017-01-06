package services

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
	assert.Equal("service", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("svc", command.Aliases[0], "Aliases should match")
	assert.Equal("options for managing services", command.Usage, "Usage should match")
	assert.Equal(4, len(command.Subcommands), "Subcommands len should match")
}