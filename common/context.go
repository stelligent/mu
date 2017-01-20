package common

import (
	"bufio"
	"bytes"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path"
	"time"
)

var version string

// GetVersion returns the current version of the app
func GetVersion() string {
	if version == "" {
		return "0.0.0-local"
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

// InitializeFromFile loads config from file
func (ctx *Context) InitializeFromFile(muFile string) error {
	// load yaml config
	yamlFile, err := os.Open(muFile)
	if err != nil {
		return err
	}
	defer func() {
		yamlFile.Close()
	}()

	repoName := path.Base(path.Dir(muFile))
	repoRevision := time.Now().Format("20060102150405")
	gitRevision, err := findGitRevision(muFile)
	if err == nil {
		repoRevision = gitRevision
	}

	return ctx.Initialize(bufio.NewReader(yamlFile), repoName, repoRevision)
}

// Initialize loads config object
func (ctx *Context) Initialize(configReader io.Reader, repoName string, repoRevision string) error {
	// initialize the repo
	ctx.Repo.Name = repoName
	ctx.Repo.Revision = repoRevision

	// load the configuration
	err := loadYamlConfig(&ctx.Config, configReader)
	if err != nil {
		return err
	}

	// service defaults
	if ctx.Config.Service.Name == "" {
		ctx.Config.Service.Name = repoName
	}
	if ctx.Config.Service.Revision == "" {
		ctx.Config.Service.Revision = repoRevision
	}

	// initialize StackManager
	ctx.StackManager, err = newStackManager(ctx.Config.Region)
	if err != nil {
		return err
	}

	// initialize ClusterManager
	ctx.ClusterManager, err = newClusterManager(ctx.Config.Region)
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
