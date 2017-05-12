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
	} else {

		// The .git folder does not exist, check to see if we are in CodeBuild
		if os.Getenv("CODEBUILD_INITIATOR") != "" {
			log.Debugf("Trying to determine git revision from CodeBuild initiator.")
			initiator := os.Getenv("CODEBUILD_INITIATOR")
			parts := strings.Split(initiator, "/")

			// See if the build was initiated by CodePipeline
			if parts[0] == "codepipeline" {
				// Try retrieving the revision from the CodePipeline status
				gitInfo, err := ctx.PipelineManager.GetGitInfo(parts[1])
				if err != nil {
					log.Warningf("Unable to determine git information from CodeBuild initiator: %s", initiator)
				}

				ctx.Config.Repo.Provider = gitInfo.provider
				ctx.Config.Repo.Revision = string(gitInfo.revision[:7])
				ctx.Config.Repo.Name = gitInfo.repoName
				ctx.Config.Repo.Slug = gitInfo.slug
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
func (ctx *Context) InitializeContext(profile string, region string, dryrun bool) error {
	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}
	if region != Empty {
		sessOptions.Config = aws.Config{Region: aws.String(region)}
	}
	if profile != Empty {
		sessOptions.Profile = profile
	}
	log.Debugf("Creating AWS session profile:%s region:%s", profile, region)
	sess, err := session.NewSessionWithOptions(sessOptions)
	if err != nil {
		return err
	}

	// initialize StackManager
	ctx.StackManager, err = newStackManager(sess, dryrun)
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

	// initialize RdsManager
	ctx.RdsManager, err = newRdsManager(sess)
	if err != nil {
		return err
	}

	// initialize ParamManager
	ctx.ParamManager, err = newParamManager(sess)
	if err != nil {
		return err
	}

	// initialize CodePipelineManager
	ctx.PipelineManager, err = newPipelineManager(sess)
	if err != nil {
		return err
	}

	// initialize CloudWatchLogs
	ctx.LogsManager, err = newLogsManager(sess)
	if err != nil {
		return err
	}

	// initialize DockerManager
	ctx.DockerManager, err = newClientDockerManager()
	if err != nil {
		return err
	}

	// initialize TaskManager
	ctx.TaskManager, err = newTaskManager(sess, dryrun)
	if err != nil {
		return err
	}

	ctx.DockerOut = os.Stdout

	return nil
}

func loadYamlConfig(config *Config, yamlReader io.Reader) error {
	yamlBuffer := new(bytes.Buffer)
	yamlBuffer.ReadFrom(yamlReader)
	return yaml.Unmarshal(yamlBuffer.Bytes(), config)
}
