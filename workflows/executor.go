package workflows

import (
	"errors"
)

// Executor define contract for the steps of a workflow
type Executor func() error

// Conditional define contract for the conditional predicate
type Conditional func() bool

func newPipelineExecutor(executors ...Executor) Executor {
	return func() error {
		for _, executor := range executors {
			err := executor()
			if err != nil {
				log.Errorf("%v", err)
				log.Debugf("%+v", err)
				return errors.New("")
			}
		}
		return nil
	}
}

func newConditionalExecutor(conditional Conditional, trueExecutor Executor, falseExecutor Executor) Executor {
	return func() error {
		if conditional() == true {
			if trueExecutor != nil {
				return trueExecutor()
			}
		} else {
			if falseExecutor != nil {
				return falseExecutor()
			}
		}
		return nil
	}
}
