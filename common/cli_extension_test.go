package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

type mockedCliPrompt struct {
	mock.Mock
}

func (m *mockedCliPrompt) Prompt(prompt string, def bool) (bool, error) {
	args := m.Called(prompt, def)
	return args.Bool(0), args.Error(1)
}

func TestPrompt_True(t *testing.T) {
	assert := assert.New(t)

	cli := new(mockedCliPrompt)
	cli.On("Prompt", "foo", true).Return(true, nil)

	answer, err := cli.Prompt("foo", true)

	assert.Nil(err)

	cli.AssertExpectations(t)
	cli.AssertNumberOfCalls(t, "Prompt", 1)

	assert.Equal(answer, true)
}
func TestPrompt_False(t *testing.T) {
	assert := assert.New(t)

	cli := new(mockedCliPrompt)
	cli.On("Prompt", "foo", true).Return(false, nil)

	answer, err := cli.Prompt("foo", true)

	assert.Nil(err)

	cli.AssertExpectations(t)
	cli.AssertNumberOfCalls(t, "Prompt", 1)

	assert.Equal(answer, false)
}
