package workflows

import (
	"encoding/base64"
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"strings"
)

type serviceWorkflow struct {
	serviceName  string
	serviceTag   string
	serviceImage string
	registryAuth string
}

// Find a service in config, by name and set the reference
func (workflow *serviceWorkflow) serviceLoader(ctx *common.Context, tag string) Executor {
	return func() error {
		// Repo Name
		if ctx.Config.Service.Name == "" {
			workflow.serviceName = ctx.Repo.Name
		} else {
			workflow.serviceName = ctx.Config.Service.Name
		}

		// Tag
		if tag != "" {
			workflow.serviceTag = tag
		} else {
			workflow.serviceTag = ctx.Repo.Revision
		}

		log.Debugf("Working with service:'%s' tag:'%s'", workflow.serviceName, workflow.serviceTag)
		return nil
	}
}

func (workflow *serviceWorkflow) serviceRepoUpserter(service *common.Service, stackUpserter common.StackUpserter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		if service.ImageRepository != "" {
			log.Noticef("Using repo '%s' for service '%s'", service.ImageRepository, workflow.serviceName)
			workflow.serviceImage = service.ImageRepository
			return nil
		}

		log.Noticef("Upsert repo for service '%s'", workflow.serviceName)

		template, err := templates.NewTemplate("repo.yml", nil)
		if err != nil {
			return err
		}

		stackParams := make(map[string]string)
		stackParams["RepoName"] = workflow.serviceName

		ecrStackName := common.CreateStackName(common.StackTypeRepo, workflow.serviceName)

		err = stackUpserter.UpsertStack(ecrStackName, template, stackParams, buildEnvironmentTags(workflow.serviceName, common.StackTypeRepo))
		if err != nil {
			return err
		}

		log.Debugf("Waiting for stack '%s' to complete", ecrStackName)
		stack := stackWaiter.AwaitFinalStatus(ecrStackName)
		workflow.serviceImage = fmt.Sprintf("%s:%s", stack.Outputs["RepoUrl"], workflow.serviceTag)
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

		workflow.registryAuth = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("{\"username\":\"%s\", \"password\":\"%s\"}", authParts[0], authParts[1])))
		return nil
	}
}
