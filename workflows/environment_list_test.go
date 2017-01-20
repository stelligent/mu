package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEnvironmentLister(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	lister := NewEnvironmentLister(ctx, nil)
	assert.NotNil(lister)
}
