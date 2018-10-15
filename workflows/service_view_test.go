package workflows

import (
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceViewer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	viewer := NewServiceViewer(ctx, "foo", nil, false)
	assert.NotNil(viewer)
}
