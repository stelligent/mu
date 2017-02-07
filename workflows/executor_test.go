package workflows

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWorkflow(t *testing.T) {
	assert := assert.New(t)

	// empty
	emptyWorkflow := newWorkflow()
	assert.Nil(emptyWorkflow())

	// error case
	errorWorkflow := newWorkflow(func() error {
		return errors.New("error occurred")
	})
	assert.NotNil(errorWorkflow())

	// multiple success case
	runcount := 0
	successWorkflow := newWorkflow(
		func() error {
			runcount = runcount + 1
			return nil
		},
		func() error {
			runcount = runcount + 1
			return nil
		})
	assert.Nil(successWorkflow())
	assert.Equal(2, runcount)
}
