package common

import (
	"testing"
	"github.com/stretchr/testify/assert"

)

func TestNewStack(t *testing.T) {
	assert := assert.New(t)

	stack := NewStack("foo")

	assert.NotNil(stack)
	assert.Equal("foo",stack.Name)
}
