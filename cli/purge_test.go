package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPurgeCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newPurgeCommand(ctx)

	assert.NotNil(command)
	assert.Equal("purge", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("nuke", command.Aliases[0], "Aliases should match")
}
