package workflows

import (
	"errors"
)

// Executor define contract for the steps of a workflow
type Executor func() error

func newWorkflow(executors ...Executor) Executor {
	return func() error {
		for _, executor := range executors {
			err := executor()
			if err != nil {
				log.Errorf("%v", err)
				return errors.New("")
			}
		}
		return nil
	}
}
