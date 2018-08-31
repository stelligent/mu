package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommand(t *testing.T) {
	// TODO: not really testing anything here yet
	os.Args = []string{"mu", "validate"}
	assert := assert.New(t)
	app := NewApp()
	app.Run(os.Args)
	assert.NotNil(app)
}
