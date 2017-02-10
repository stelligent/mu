package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEnvironmentViewer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	viewer := NewEnvironmentViewer(ctx, "json", "foo", nil)
	assert.NotNil(viewer)
}
