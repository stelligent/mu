package workflows

import (
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewEnvironmentViewer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	viewer := NewEnvironmentViewer(ctx, "json", "foo", nil)
	assert.NotNil(viewer)
}
