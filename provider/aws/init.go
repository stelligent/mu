package aws

import (
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stelligent/mu/common"
)

// InitializeContext loads manager objects
func InitializeContext(ctx *common.Context, profile string, assumeRole string, region string, dryrunPath string, skipVersionCheck bool, proxy string) error {

	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}

	if _, present := os.LookupEnv("MU_AWS_HTTP_DEBUG"); present {
		sessOptions.Config.WithLogLevel(aws.LogDebugWithHTTPBody)
	}
	if region != common.Empty {
		sessOptions.Config.Region = aws.String(region)
	}
	if proxy != common.Empty {
		proxyHTTPClient := &http.Client{
			Transport: &http.Transport{
				Proxy: func(*http.Request) (*url.URL, error) {
					return url.Parse(proxy)
				},
			},
		}
		sessOptions.Config.HTTPClient = proxyHTTPClient
	}
	if profile != common.Empty {
		sessOptions.Profile = profile
	}

	log.Debugf("Creating AWS session profile:%s region:%s proxy:%s", profile, region, proxy)
	sess, err := session.NewSessionWithOptions(sessOptions)
	if err != nil {
		return err
	}

	if assumeRole != common.Empty {
		// Create the credentials from AssumeRoleProvider to assume the role
		// referenced by the "myRoleARN" ARN.
		creds := stscreds.NewCredentials(sess, assumeRole)
		sess, err = session.NewSession(&aws.Config{Region: sess.Config.Region, HTTPClient: sess.Config.HTTPClient, Credentials: creds})
		if err != nil {
			return err
		}
	}

	// initialize StackManager
	ctx.StackManager, err = newStackManager(sess, ctx.ExtensionsManager, dryrunPath, skipVersionCheck)
	if err != nil {
		return err
	}

	// initialize ClusterManager
	ctx.ClusterManager, err = newClusterManager(sess)
	if err != nil {
		return err
	}

	// initialize InstanceManager
	ctx.InstanceManager, err = newInstanceManager(sess)
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

	// initialize TaskManager
	ctx.TaskManager, err = newTaskManager(sess, &ctx.StackManager)
	if err != nil {
		return err
	}

	// initialize ArtifactManager
	ctx.ArtifactManager, err = newArtifactManager(sess)
	if err != nil {
		return err
	}

	// initialize SubscriptionManager
	ctx.SubscriptionManager, err = newSnsManager(sess)
	if err != nil {
		return err
	}

	// initialize the RolesetManager
	ctx.RolesetManager, err = newRolesetManager(ctx)
	if err != nil {
		return err
	}

	// initialize LocalCodePipelineManager
	localSess, err := session.NewSession()
	if err != nil {
		return err
	}
	ctx.LocalPipelineManager, _ = newPipelineManager(localSess)

	ctx.DockerOut = os.Stdout

	return nil
}
