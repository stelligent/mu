package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stelligent/mu/common"
)

type ecsClusterManager struct {
	ecsAPI ecsiface.ECSAPI
	ecrAPI ecriface.ECRAPI
	dryrun bool
}

func newClusterManager(sess *session.Session, dryrun bool) (common.ClusterManager, error) {
	log.Debug("Connecting to ECS service")
	ecsAPI := ecs.New(sess)

	log.Debug("Connecting to ECR service")
	ecrAPI := ecr.New(sess)

	return &ecsClusterManager{
		ecsAPI: ecsAPI,
		ecrAPI: ecrAPI,
		dryrun: dryrun,
	}, nil
}

// ListInstances get the instances for a specific cluster
func (ecsMgr *ecsClusterManager) ListInstances(clusterName string) ([]common.ContainerInstance, error) {
	ecsAPI := ecsMgr.ecsAPI

	params := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	}

	log.Debugf("Searching for container instances for cluster named '%s'", clusterName)

	var instanceIds []*string
	err := ecsAPI.ListContainerInstancesPages(params, func(page *ecs.ListContainerInstancesOutput, lastPage bool) bool {
		for _, instanceID := range page.ContainerInstanceArns {
			instanceIds = append(instanceIds, instanceID)
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	describeParams := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(clusterName),
		ContainerInstances: instanceIds,
	}
	describeOut, _ := ecsAPI.DescribeContainerInstances(describeParams)

	instances := make([]common.ContainerInstance, len(describeOut.ContainerInstances))
	for i, instance := range describeOut.ContainerInstances {
		instances[i] = instance
	}
	return instances, nil
}

func (ecsMgr *ecsClusterManager) AuthenticateRepository(repoURL string) (string, error) {
	ecrAPI := ecsMgr.ecrAPI

	params := &ecr.GetAuthorizationTokenInput{}

	log.Debug("Authenticating to ECR repo")

	resp, err := ecrAPI.GetAuthorizationToken(params)
	if err != nil {
		return common.Empty, err
	}

	for _, authData := range resp.AuthorizationData {
		if strings.HasPrefix(fmt.Sprintf("https://%s", repoURL), aws.StringValue(authData.ProxyEndpoint)) {
			return aws.StringValue(authData.AuthorizationToken), nil
		}
	}

	return common.Empty, fmt.Errorf("unable to find token for repo url:%s", repoURL)
}

func (ecsMgr *ecsClusterManager) DeleteRepository(repoName string) error {
	ecrAPI := ecsMgr.ecrAPI

	if ecsMgr.dryrun {
		log.Infof("  DRYRUN: Skipping deletion of repository '%s'", repoName)
		return nil
	}
	log.Infof("  Deleting repository '%s'", repoName)
	ecrAPI.DeleteRepository(&ecr.DeleteRepositoryInput{
		Force:          aws.Bool(true),
		RepositoryName: aws.String(repoName),
	})
	return nil
}
