package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewServiceDeployer(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	deploye := NewServiceDeployer(ctx, "dev", "foo")
	assert.NotNil(deploye)
}

/*
func TestServiceDeploy(t *testing.T) {
	assert := assert.New(t)

	workflow := new(serviceWorkflow)
	workflow.serviceName = "foo"

	err := workflow.serviceDeployer("dev")()
	assert.Nil(err)
}
*/
