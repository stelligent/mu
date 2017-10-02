package common

import (
	"bufio"
	"bytes"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var version string

// GetVersion returns the current version of the app
func GetVersion() string {
	if version == Empty {
		return DefaultVersion
	}
	return version
}

// SetVersion returns the current version of the app
func SetVersion(v string) {
	version = v
}

// NewContext create a new context object
func NewContext() *Context {
	ctx := new(Context)
	return ctx
}

// InitializeConfigFromFile loads config from file
func (ctx *Context) InitializeConfigFromFile(muFile string) error {
	absMuFile, err := filepath.Abs(muFile)

	// set the basedir
	ctx.Config.Basedir = path.Dir(absMuFile)
	log.Debugf("Setting basedir=%s", ctx.Config.Basedir)

	ctx.Config.Repo.Name = path.Base(ctx.Config.Basedir)
	ctx.Config.Repo.Revision = time.Now().Format("20060102150405")

	ctx.Config.RelMuFile, err = getRelMuFile(absMuFile)
	if err != nil {
		return err
	}

	// Get the git revision from the .git folder
	gitRevision, err := findGitRevision(ctx.Config.Basedir)
	if err == nil {
		ctx.Config.Repo.Revision = gitRevision
		gitURL, err := findGitRemoteURL(ctx.Config.Basedir)
		if err == nil {
			gitProvider, gitSlug, err := findGitSlug(gitURL)
			if err == nil {
				ctx.Config.Repo.Provider = gitProvider
				ctx.Config.Repo.Slug = gitSlug
			} else {
				log.Warningf("Unable to determine git slug: %s", err.Error())
			}
		} else {
			log.Warningf("Unable to determine git remote url: %s", err.Error())
		}

		gitBranch, err := findGitBranch(ctx.Config.Basedir)
		if err == nil {
			ctx.Config.Repo.Branch = gitBranch
		} else {
			log.Warningf("Unable to determine git branch: %s", err.Error())
		}
	} else {

		// The .git folder does not exist, check to see if we are in CodeBuild
		if os.Getenv("CODEBUILD_INITIATOR") != "" {
			log.Debugf("Trying to determine git revision from CodeBuild initiator.")
			initiator := os.Getenv("CODEBUILD_INITIATOR")
			parts := strings.Split(initiator, "/")

			// See if the build was initiated by CodePipeline
			if parts[0] == "codepipeline" {
				// Try retrieving the revision from the CodePipeline status
				gitInfo, err := ctx.LocalPipelineManager.GetGitInfo(parts[1])
				if err != nil {
					log.Warningf("Unable to determine git information from CodeBuild initiator '%s': %s", initiator, err)
				}

				sourceVersion := os.Getenv("CODEBUILD_RESOLVED_SOURCE_VERSION")
				if sourceVersion == "" {
					sourceVersion = gitInfo.Revision
				}
				if len(sourceVersion) > 7 {
					ctx.Config.Repo.Revision = string(sourceVersion[:7])
				}

				ctx.Config.Repo.Name = gitInfo.RepoName
				ctx.Config.Repo.Slug = gitInfo.Slug
				ctx.Config.Repo.Provider = gitInfo.Provider
			} else {
				log.Warningf("Unable to process CodeBuild initiator: %s", initiator)
			}
		} else {
			log.Warningf("Unable to determine git revision: %s", err.Error())
		}
	}
	log.Debugf("Setting repo provider=%s", ctx.Config.Repo.Provider)
	log.Debugf("Setting repo name=%s", ctx.Config.Repo.Name)
	log.Debugf("Setting repo revision=%s", ctx.Config.Repo.Revision)
	log.Debugf("Setting repo slug=%s", ctx.Config.Repo.Slug)

	// load yaml config
	yamlFile, err := os.Open(absMuFile)
	if err != nil {
		return err
	}
	defer func() {
		yamlFile.Close()
	}()

	return ctx.InitializeConfig(bufio.NewReader(yamlFile))
}

func getRelMuFile(absMuFile string) (string, error) {
	var repoDir string
	gitDir, error := findGitDirectory(absMuFile)
	if error != nil {
		repoDir, error = os.Getwd()
		if error != nil {
			return "", error
		}
	} else {
		repoDir = filepath.Dir(gitDir)
	}

	absRepoDir, error := filepath.Abs(repoDir)
	if error != nil {
		return "", error
	}

	relMuFile, error := filepath.Rel(absRepoDir, absMuFile)
	if error != nil {
		return "", error
	}

	log.Debugf("Absolute repodir: %s", absRepoDir)
	log.Debugf("Absolute mu file: %s", absMuFile)
	log.Debugf("Relative mu file: %s", relMuFile)

	return relMuFile, nil
}

// InitializeConfig loads config object
func (ctx *Context) InitializeConfig(configReader io.Reader) error {

	// load the configuration
	err := loadYamlConfig(&ctx.Config, configReader)
	if err != nil {
		return err
	}

	// register the stack overrides
	registerStackOverrides(ctx.Config.Templates)

	return nil
}

// InitializeContext loads manager objects
func (ctx *Context) InitializeContext() error {
	var err error

	// initialize DockerManager
	ctx.DockerManager, err = newClientDockerManager()
	if err != nil {
		return err
	}

	return nil
}

func loadYamlConfig(config *Config, yamlReader io.Reader) error {
	yamlBuffer := new(bytes.Buffer)
	yamlBuffer.ReadFrom(yamlReader)
	return yaml.Unmarshal(yamlBuffer.Bytes(), config)
}
