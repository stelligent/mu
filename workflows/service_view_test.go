package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServiceViewer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	viewer := NewServiceViewer(ctx, "foo", nil)
	assert.NotNil(viewer)
}
