package services

import (
	"testing"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewUndeployCommand(t *testing.T) {
	assert := assert.New(t)

	config := &common.Config {}

	command := newUndeployCommand(config)

	assert.NotNil(command)
	assert.Equal("undeploy", command.Name, "Name should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.Equal(1, len(command.Flags), "Flags length")
	assert.Equal("service, s", command.Flags[0].GetName(), "Flags Name")
	assert.NotNil(command.Action)
}

