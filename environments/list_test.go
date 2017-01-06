package environments

import (
	"testing"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewListCommand(t *testing.T) {
	assert := assert.New(t)

	config := &common.Config {}

	command := newListCommand(config)

	assert.NotNil(command)
	assert.Equal("list", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("ls", command.Aliases[0], "Aliases should match")
	assert.Equal("list environments", command.Usage, "Usage should match")
	assert.NotNil(command.Action)
}