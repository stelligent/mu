package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceLoader(t *testing.T) {
	assert := assert.New(t)

	ctx := new(common.Context)
	ctx.Config.Service.Name = "myservice"

	workflow := new(serviceWorkflow)
	err := workflow.serviceLoader(ctx, "foo")()
	assert.Nil(err)
	assert.Equal("myservice", workflow.serviceName)
}
