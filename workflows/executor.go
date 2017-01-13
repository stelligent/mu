package workflows

import (
	"errors"
	"github.com/op/go-logging"
)

// Executor define contract for the steps of a workflow
type Executor func() error

func newWorkflow(executors ...Executor) Executor {
	var log = logging.MustGetLogger("workflow")
	return func() error {
		for _, executor := range executors {
			err := executor()
			if err != nil {
				log.Error(err)
				return errors.New("")
			}
		}
		return nil
	}
}
