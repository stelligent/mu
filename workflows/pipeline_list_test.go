package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPipelineLister(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	lister := NewPipelineLister(ctx, nil)
	assert.NotNil(lister)
}
