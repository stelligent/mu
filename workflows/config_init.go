package workflows

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	"gopkg.in/yaml.v2"
)

// NewConfigInitializer create a new mu.yml file
func NewConfigInitializer(ctx *common.Context, createEnvironment bool, listenPort int, forceOverwrite bool) Executor {

	workflow := new(configWorkflow)

	return newPipelineExecutor(
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

		// write config
		config.Service.Port = listenPort
		config.Service.PathPatterns = []string{"/*"}

		if createEnvironment && len(config.Environments) == 0 {
			config.Environments = append(config.Environments,
				common.Environment{Name: "acceptance"},
				common.Environment{Name: "production"})
		}

		if config.Namespace == "mu" {
			config.Namespace = ""
		}

		configBytes, err := yaml.Marshal(config)
		if err != nil {
			return err
		}

		// unless force is set, don't overwrite...make sure files don't exist
		if _, err := os.Stat(fmt.Sprintf("%s/mu.yml", basedir)); !forceOverwrite && err == nil {
			log.Infof("Config file already exists - '%s/mu.yml'.  Use --force to overwrite", basedir)
		} else {
			log.Noticef("Writing config to '%s/mu.yml'", basedir)
			err = ioutil.WriteFile(fmt.Sprintf("%s/mu.yml", basedir), configBytes, 0600)
			if err != nil {
				return err
			}
		}

		// write buildspec
		buildspec, err := templates.GetAsset(common.TemplateBuildspec,
			templates.ExecuteTemplate(nil))
		if err != nil {
			return err
		}
		buildspecBytes := []byte(buildspec)

		for _, path := range []string{
			"buildspec.yml", "buildspec-test.yml", "buildspec-prod.yml",
		} {
			abspath := fmt.Sprintf("%s/%s", basedir, path)
			if _, err := os.Stat(abspath); !forceOverwrite && err == nil {
				log.Infof("buildspec file already exists - '%s'.  Use --force to overwrite", abspath)
				continue
			}

			log.Noticef("Writing buildspec to '%s'", abspath)
			err = ioutil.WriteFile(fmt.Sprintf("%s", abspath), buildspecBytes, 0600)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
