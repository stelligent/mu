package workflows

import (
	"errors"

	"github.com/stelligent/mu/common"
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
				switch err.(type) {
				case common.Warning:
					log.Warning(err.Error())
					return nil
				default:
					log.Errorf("%v", err)
					log.Debugf("%+v", err)
					return errors.New("")
				}
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

func executeWithChan(executor Executor, errChan chan error) {
	errChan <- executor()
}

func newErrorExecutor(err error) Executor {
	return func() error {
		return err
	}
}

func newParallelExecutor(executors ...Executor) Executor {
	return func() error {
		errChan := make(chan error)

		for _, executor := range executors {
			go executeWithChan(executor, errChan)
		}

		for i := 0; i < len(executors); i++ {
			err := <-errChan
			if err != nil {
				return err
			}
		}
		return nil
	}
}
