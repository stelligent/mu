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

func TestServiceDeploy(t *testing.T) {
	assert := assert.New(t)

	workflow := new(serviceWorkflow)
	workflow.service = &common.Service{
		Name: "foo",
	}

	err := workflow.serviceDeployer("dev", "foo")()
	assert.Nil(err)
}
