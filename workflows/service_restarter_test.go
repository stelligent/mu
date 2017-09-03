package workflows

import (
	"testing"

	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceRestarter(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	restarter := NewServiceRestarter(ctx, "foo", "foo", 0)
	assert.NotNil(restarter)
}
