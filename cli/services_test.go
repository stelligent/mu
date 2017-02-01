package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServicesCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newServicesCommand(ctx)

	assert.NotNil(command)
	assert.Equal("service", command.Name, "Name should match")
	assert.Equal(1, len(command.Aliases), "Aliases len should match")
	assert.Equal("svc", command.Aliases[0], "Aliases should match")
	assert.Equal("options for managing services", command.Usage, "Usage should match")
	assert.Equal(4, len(command.Subcommands), "Subcommands len should match")
}

func TestNewServicesShowCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newServicesShowCommand(ctx)

	assert.NotNil(command)
	assert.Equal("show", command.Name, "Name should match")
	assert.Equal(0, len(command.Flags), "Flags length")
	assert.Equal("[<service>]", command.ArgsUsage, "ArgsUsage should match")
	assert.NotNil(command.Action)
}

func TestNewServicesPushCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newServicesPushCommand(ctx)

	assert.NotNil(command)
	assert.Equal("push", command.Name, "Name should match")
	assert.Equal(1, len(command.Flags), "Flags length")
	assert.Equal("tag, t", command.Flags[0].GetName(), "Flags Name")
	assert.NotNil(command.Action)
}

func TestNewServicesDeployCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newServicesDeployCommand(ctx)

	assert.NotNil(command)
	assert.Equal("deploy", command.Name, "Name should match")
	assert.Equal("<environment>", command.ArgsUsage, "ArgsUsage should match")
	assert.Equal(1, len(command.Flags), "Flags length")
	assert.Equal("tag, t", command.Flags[0].GetName(), "Flags Name")
	assert.NotNil(command.Action)
}

func TestNewUndeployCommand(t *testing.T) {
	assert := assert.New(t)

	ctx := common.NewContext()

	command := newServicesUndeployCommand(ctx)

	assert.NotNil(command)
	assert.Equal("undeploy", command.Name, "Name should match")
	assert.Equal("<environment> [<service>]", command.ArgsUsage, "ArgsUsage should match")
	assert.Equal(0, len(command.Flags), "Flags length")
	assert.NotNil(command.Action)
}
