package common

import (
	"testing"
	"github.com/stretchr/testify/assert"

)

func TestNewStack(t *testing.T) {
	assert := assert.New(t)

	stack := NewStack("foo","us-west-2")

	assert.NotNil(stack)
	assert.Equal("foo",stack.Name)
}
