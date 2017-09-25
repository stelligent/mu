package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stelligent/mu/common"
	"os"
)

// InitializeContext loads manager objects
func InitializeContext(ctx *common.Context, profile string, region string, dryrun bool) error {
	sessOptions := session.Options{SharedConfigState: session.SharedConfigEnable}
	if region != common.Empty {
		sessOptions.Config = aws.Config{Region: aws.String(region)}
	}
	if profile != common.Empty {
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

	// initialize the RolesetManager
	ctx.RolesetManager, err = newRolesetManager(ctx)
	if err != nil {
		return err
	}

	ctx.DockerOut = os.Stdout

	return nil
}
