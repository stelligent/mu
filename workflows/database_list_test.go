package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDatabaseLister(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	lister := NewDatabaseLister(ctx, nil)
	assert.NotNil(lister)
}
