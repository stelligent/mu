package common

import (
	"testing"
	"github.com/stretchr/testify/assert"

)

func TestNewEnvironmentStack(t *testing.T) {
	assert := assert.New(t)

	env := Environment{
		Name: "test",
	}

	stack, err := env.NewStack()

	assert.Nil(err)
	assert.NotNil(stack)
	assert.Equal("mu-env-test",stack.Name)
}
