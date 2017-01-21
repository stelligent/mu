package workflows

import (
	"github.com/stelligent/mu/common"
	"encoding/base64"
	"strings"
	"fmt"
)

type serviceWorkflow struct {
	serviceName string
	serviceImage string
	serviceTag string
	registryAuth string
}

// Find a service in config, by name and set the reference
func (workflow *serviceWorkflow) serviceLoader(config *common.Config, tag string) Executor {
	return func() error {
		workflow.serviceName = config.Service.Name

		if tag != "" {
			workflow.serviceTag = tag
		} else {
			workflow.serviceTag = config.Service.Revision
		}

		log.Debugf("Working with service:'%s' tag:'%s'", workflow.serviceName, tag)
		return nil
	}
}

func (workflow *serviceWorkflow) serviceRegistryAuthenticator(authenticator common.RepositoryAuthenticator) Executor {
	return func() error {
		log.Debugf("Authenticating to registry '%s'", workflow.serviceImage)
		registryAuth, err := authenticator.AuthenticateRepository(workflow.serviceImage)
		if err != nil {
			return err
		}

		data, err := base64.StdEncoding.DecodeString(registryAuth)
		if err != nil {
			return err
		}

		authParts := strings.Split(string(data), ":")

		workflow.registryAuth = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("{\"username\":\"%s\", \"password\":\"%s\"}",authParts[0],authParts[1])))
		return nil
	}
}
