package workflows

import (
	"github.com/stelligent/mu/common"
)

type serviceWorkflow struct {
	service *common.Service
}

// Find a service in config, by name and set the reference
func (workflow *serviceWorkflow) serviceLoader(config *common.Config) Executor {
	return func() error {
		workflow.service = &config.Service

		log.Debugf("Working with service '%s' version '%s'", workflow.service.Name, workflow.service.Revision)
		return nil
	}
}
