package cli

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	assert := assert.New(t)
	app := NewApp("1.2.3")

	assert.NotNil(app)
	assert.Equal("mu", app.Name, "Name should match")
	assert.Equal("1.2.3", app.Version, "Version should match")
	assert.Equal("Microservice Platform on AWS", app.Usage, "usage should match")
	assert.Equal(true, app.EnableBashCompletion, "bash completion should match")
	assert.Equal(1, len(app.Flags), "Flags len should match")
	assert.Equal("config, c", app.Flags[0].GetName(), "Flags name should match")
	assert.Equal(3, len(app.Commands), "Commands len should match")
	assert.Equal("environment", app.Commands[0].Name, "Command[0].name should match")
	assert.Equal("service", app.Commands[1].Name, "Command[1].name should match")
	assert.Equal("pipeline", app.Commands[2].Name, "Command[2].name should match")
}

