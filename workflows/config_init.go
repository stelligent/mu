package workflows

import (
	"bytes"
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
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

		if config.Repo.Slug == "" {
			return fmt.Errorf("Unable to determine git repo to use for the pipeline.  Have you initialized your repo and pushed yet? %s", config.Repo.Slug)
		}

		// unless force is set, don't overwrite...make sure files don't exist
		if forceOverwrite == false {
			log.Debugf("Checking for existing config file at %s/mu.yml", basedir)
			if _, err := os.Stat(fmt.Sprintf("%s/mu.yml", basedir)); err == nil {
				return fmt.Errorf("Config file already exists - '%s/mu.yml'.  Use --force to overwrite", basedir)
			}

			log.Debugf("Checking for existing buildspec file at %s/buildspec.yml", basedir)
			if _, err := os.Stat(fmt.Sprintf("%s/buildspec.yml", basedir)); err == nil {
				return fmt.Errorf("buildspec file already exists - '%s/buildspec.yml'.  Use --force to overwrite", basedir)
			}
		}

		// write config
		config.Service.Port = listenPort
		config.Service.Name = config.Repo.Name
		config.Service.PathPatterns = []string{"/*"}
		config.Service.Pipeline.Source.Repo = config.Repo.Slug
		config.Service.Pipeline.Source.Provider = config.Repo.Provider

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

		// write buildspec
		buildspec, err := templates.NewTemplate("buildspec.yml", nil, nil)
		if err != nil {
			return err
		}
		buildspecBytes := new(bytes.Buffer)
		buildspecBytes.ReadFrom(buildspec)

		log.Noticef("Writing buildspec to '%s/buildspec.yml'", basedir)

		err = ioutil.WriteFile(fmt.Sprintf("%s/buildspec.yml", basedir), buildspecBytes.Bytes(), 0600)
		if err != nil {
			return err
		}

		return nil
	}
}
