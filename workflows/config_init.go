package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// NewConfigInitializer create a new mu.yml file
func NewConfigInitializer(ctx *common.Context, createEnvironment bool, listenPort int, forceOverwrite bool) Executor {

	workflow := new(configWorkflow)

	return newWorkflow(
		workflow.configInitialize(&ctx.Config, createEnvironment, listenPort, forceOverwrite),
	)
}

type configWorkflow struct {
}

func (workflow *configWorkflow) configInitialize(config *common.Config, createEnvironment bool, listenPort int, forceOverwrite bool) Executor {
	return func() error {
		basedir := "."
		if config.Basedir != "" {
			basedir = config.Basedir
		}

		if forceOverwrite == false {
			log.Debugf("Checking for existing config file at %s/mu.yml", basedir)
			// don't overwrite...make sure file doesn't exist
			if _, err := os.Stat(fmt.Sprintf("%s/mu.yml", basedir)); err == nil {
				return fmt.Errorf("Config file already exists - '%s/mu.yml'.  Use --force to overwrite", basedir)
			}
		}

		config.Service.Port = listenPort
		config.Service.Name = config.Repo.Name
		config.Service.PathPatterns = []string{"/*"}
		config.Service.Pipeline.Source.Repo = config.Repo.Slug

		if createEnvironment && len(config.Environments) == 0 {
			config.Environments = append(config.Environments,
				common.Environment{Name: "dev"},
				common.Environment{Name: "production"})
		}

		configBytes, err := yaml.Marshal(config)
		if err != nil {
			return err
		}

		log.Noticef("Writing config to '%s/mu.yml'", basedir)

		err = ioutil.WriteFile(fmt.Sprintf("%s/mu.yml", basedir), configBytes, 0600)
		if err != nil {
			return err
		}

		return nil
	}
}
