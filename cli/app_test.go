package cli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewApp(t *testing.T) {
	assert := assert.New(t)
	app := NewApp()

	assert.NotNil(app)
	assert.Equal("mu", app.Name, "Name should match")
	assert.Equal("1.0.0-local", app.Version, "Version should match")
	assert.Equal("Microservice Platform on AWS", app.Usage, "usage should match")
	assert.Equal(true, app.EnableBashCompletion, "bash completion should match")
	assert.Equal(11, len(app.Flags), "Flags len should match")
	assert.Equal("config, c", app.Flags[0].GetName(), "Flags name should match")
	assert.Equal("region, r", app.Flags[1].GetName(), "Flags name should match")
	assert.Equal("assume-role, a", app.Flags[2].GetName(), "Flags name should match")
	assert.Equal("profile, p", app.Flags[3].GetName(), "Flags name should match")
	assert.Equal("namespace, n", app.Flags[4].GetName(), "Flags name should match")
	assert.Equal("silent, s", app.Flags[5].GetName(), "Flags name should match")
	assert.Equal("verbose, V", app.Flags[6].GetName(), "Flags name should match")
	assert.Equal("dryrun, d", app.Flags[7].GetName(), "Flags name should match")
	assert.Equal("dryrun-output, O", app.Flags[8].GetName(), "Flags name should match")
	assert.Equal("disable-iam, I", app.Flags[9].GetName(), "Flags name should match")
	assert.Equal("skip-version-check, F", app.Flags[10].GetName(), "Flags name should match")
	assert.Equal(5, len(app.Commands), "Commands len should match")
	assert.Equal("init", app.Commands[0].Name, "Command[0].name should match")
	assert.Equal("environment", app.Commands[1].Name, "Command[1].name should match")
	assert.Equal("service", app.Commands[2].Name, "Command[2].name should match")
	assert.Equal("pipeline", app.Commands[3].Name, "Command[3].name should match")
}
