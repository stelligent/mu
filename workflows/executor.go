package workflows

// Executor define contract for the steps of a workflow
type Executor func() error

func newWorkflow(exectors ...Executor) Executor {
	return func() error {
		for _, executor := range exectors {
			err := executor()
			if err != nil {
				return err
			}
		}
		return nil
	}
}
