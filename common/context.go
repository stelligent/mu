package common

import (
	"bufio"
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path"
	"path/filepath"
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

// InitializeConfigFromFile loads config from file
func (ctx *Context) InitializeConfigFromFile(muFile string) error {
	absMuFile, err := filepath.Abs(muFile)
	if err != nil {
		return err
	}

	// load yaml config
	yamlFile, err := os.Open(absMuFile)
	if err != nil {
		return err
	}
	defer func() {
		yamlFile.Close()
	}()

	// set the basedir
	ctx.Config.Basedir = path.Dir(absMuFile)
	ctx.Config.Repo.Name = path.Base(ctx.Config.Basedir)
	ctx.Config.Repo.Revision = time.Now().Format("20060102150405")
	gitRevision, err := findGitRevision(absMuFile)
	if err == nil {
		ctx.Config.Repo.Revision = gitRevision
	}

	return ctx.InitializeConfig(bufio.NewReader(yamlFile))
}

// InitializeConfig loads config object
func (ctx *Context) InitializeConfig(configReader io.Reader) error {

	// load the configuration
	err := loadYamlConfig(&ctx.Config, configReader)
	if err != nil {
		return err
	}
	return nil
}

// InitializeContext loads manager objects
func (ctx *Context) InitializeContext(profile string, region string) error {
	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}
	if region != "" {
		sessOptions.Config = aws.Config{Region: aws.String(region)}
	}
	if profile != "" {
		sessOptions.Profile = profile
	}
	log.Debugf("Creating AWS session profile:%s region:%s", profile, region)
	sess, err := session.NewSessionWithOptions(sessOptions)
	if err != nil {
		return err
	}

	// initialize StackManager
	ctx.StackManager, err = newStackManager(sess)
	if err != nil {
		return err
	}

	// initialize ClusterManager
	ctx.ClusterManager, err = newClusterManager(sess)
	if err != nil {
		return err
	}

	// initialize ElbManager
	ctx.ElbManager, err = newElbv2Manager(sess)
	if err != nil {
		return err
	}

	// initialize CodePipelineManager
	ctx.PipelineManager, err = newPipelineManager(sess)
	if err != nil {
		return err
	}

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
