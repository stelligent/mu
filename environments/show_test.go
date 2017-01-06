package environments

import (
	"testing"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewShowCommand(t *testing.T) {
	assert := assert.New(t)

	config := &common.Config {}

	command := newShowCommand(config)

	assert.NotNil(command)
	assert.Equal("show", command.Name, "Name should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}

