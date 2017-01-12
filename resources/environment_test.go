package resources

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stelligent/mu/common"
)

func TestEnvironment_UpsertEnvironmentUnknown(t *testing.T) {
	assert := assert.New(t)
	ctx := common.NewContext()
	envMgr := NewEnvironmentManager(ctx)

	err := envMgr.UpsertEnvironment("test-stack-that-is-fake")

	assert.NotNil(err)
}

func TestGetEnvironment(t *testing.T) {
	assert := assert.New(t)
	envMgr := new(environmentManagerImpl)
	envMgr.context = common.NewContext()

	env1 := common.Environment{
		Name: "foo",
	}
	env2 := common.Environment{
		Name: "bar",
	}
	envMgr.context.Config.Environments = []common.Environment{env1, env2}

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
