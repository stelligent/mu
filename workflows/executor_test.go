package workflows

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWorkflow(t *testing.T) {
	assert := assert.New(t)

	// empty
	emptyWorkflow := newPipelineExecutor()
	assert.Nil(emptyWorkflow())

	// error case
	errorWorkflow := newPipelineExecutor(func() error {
		return errors.New("error occurred")
	})
	assert.NotNil(errorWorkflow())

	// multiple success case
	runcount := 0
	successWorkflow := newPipelineExecutor(
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

func TestNewConditionalExecutor(t *testing.T) {
	assert := assert.New(t)

	trueCount := 0
	falseCount := 0

	err := newConditionalExecutor(func() bool {
		return false
	}, func() error {
		trueCount++
		return nil
	}, func() error {
		falseCount++
		return nil
	})()

	assert.Nil(err)
	assert.Equal(0, trueCount)
	assert.Equal(1, falseCount)

	err = newConditionalExecutor(func() bool {
		return true
	}, func() error {
		trueCount++
		return nil
	}, func() error {
		falseCount++
		return nil
	})()

	assert.Nil(err)
	assert.Equal(1, trueCount)
	assert.Equal(1, falseCount)
}
