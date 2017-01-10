package resources

import (
	"testing"
	"github.com/stretchr/testify/assert"

	"github.com/stelligent/mu/common"
)

func TestEnvironment_UpsertEnvironmentUnknown(t *testing.T) {
	assert := assert.New(t)
	config := common.NewConfig()
	envMgr := NewEnvironmentManager(config)

	err := envMgr.UpsertEnvironment("test-stack-that-is-fake")

	assert.NotNil(err)
}

func TestGetEnvironment(t *testing.T) {
	assert := assert.New(t)
	config := common.NewConfig()

	env1 := common.Environment{
		Name: "foo",
	}
	env2 := common.Environment{
		Name: "bar",
	}
	config.Environments = []common.Environment{env1, env2}
	envMgr := &environmentManagerContext{
		config: config,
	}

	fooEnv, fooErr := envMgr.getEnvironment("foo")
	assert.Equal("foo", fooEnv.Name)
	assert.Nil(fooErr)

	barEnv, barErr := envMgr.getEnvironment("BAR")
	assert.Equal("bar", barEnv.Name)
	assert.Nil(barErr)

	bazEnv, bazErr := envMgr.getEnvironment("baz")
	assert.Nil(bazEnv)
	assert.NotNil(bazErr)
}
