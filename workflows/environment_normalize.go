package workflows

import (
	"fmt"

	"github.com/stelligent/mu/common"
)

// Function to normalize a workflow's environment
func (workflow *environmentWorkflow) environmentNormalizer() Executor {
	return func() error {
		if workflow.environment.Provider == "" {
			workflow.environment.Provider = common.EnvProviderEcs
		}
		if workflow.environment.Discovery.Provider == "consul" {
			return fmt.Errorf("Consul is no longer supported as a service discovery provider.  Check out the mu-consul extension for an alternative: https://github.com/stelligent/mu-consul")
		}
		return nil
	}
}
