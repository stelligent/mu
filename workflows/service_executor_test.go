package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEnvironmentExecutor(t *testing.T) {
	assertion := assert.New(t)
	ctx := common.NewContext()
	task := common.Task{
		Environment: common.TestEnv,
		Service:     common.TestSvc,
		Command:     common.TestCmd,
	}
	executor := NewServiceExecutor(ctx, task)
	assertion.NotNil(executor)
}
