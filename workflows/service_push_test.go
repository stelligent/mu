package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"testing"
)

func TestNewServicePusher(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	upserter := NewServicePusher(ctx, "foo", os.Stdout)
	assert.NotNil(upserter)
}

type mockServiceBuilder struct {
	mock.Mock
	common.DockerImageBuilder
}

func (m *mockServiceBuilder) ImageBuild(basedir string, dockerfile string, tags []string, dockerWriter io.Writer) error {
	args := m.Called()
	return args.Error(0)
}

type mockServicePusher struct {
	mock.Mock
	common.DockerImagePusher
}

func (m *mockServicePusher) ImagePush(image string, registryAuth string, dockerWriter io.Writer) error {
	args := m.Called()
	return args.Error(0)
}

func TestServiceBuilder(t *testing.T) {
	assert := assert.New(t)

	builder := new(mockServiceBuilder)
	builder.On("ImageBuild").Return(nil)

	config := new(common.Config)

	workflow := new(serviceWorkflow)
	err := workflow.serviceBuilder(builder, config, os.Stdout)()
	assert.Nil(err)

	builder.AssertExpectations(t)
	builder.AssertNumberOfCalls(t, "ImageBuild", 1)

}

func TestServicePusher(t *testing.T) {
	assert := assert.New(t)

	pusher := new(mockServicePusher)
	pusher.On("ImagePush").Return(nil)

	workflow := new(serviceWorkflow)
	err := workflow.servicePusher(pusher, os.Stdout)()
	assert.Nil(err)

	pusher.AssertExpectations(t)
	pusher.AssertNumberOfCalls(t, "ImagePush", 1)
}
