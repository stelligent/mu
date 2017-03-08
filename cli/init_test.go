package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewInitCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newInitCommand(ctx)

	assert.NotNil(command)
	assert.Equal("init", command.Name, "Name should match")
	assert.Equal("initialize mu.yml file", command.Usage, "Usage should match")
	assert.Equal(3, len(command.Flags), "Flags len should match")

}
