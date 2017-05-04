package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEnvironmentExecutor(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	executor := NewEnvironmentExecutor(ctx, "env", "cmd")
	assertion.NotNil(executor)
}
