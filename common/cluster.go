package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"strings"
)

// ClusterInstanceLister for getting cluster instances
type ClusterInstanceLister interface {
	ListInstances(clusterName string) ([]*ecs.ContainerInstance, error)
}

// RepositoryAuthenticator auths for a repo
type RepositoryAuthenticator interface {
	AuthenticateRepository(repoURL string) (string, error)
}

// ClusterManager composite of all cluster capabilities
type ClusterManager interface {
	ClusterInstanceLister
	RepositoryAuthenticator
}

type ecsClusterManager struct {
	ecsAPI ecsiface.ECSAPI
	ecrAPI ecriface.ECRAPI
}

func newClusterManager(sess *session.Session) (ClusterManager, error) {
	log.Debug("Connecting to ECS service")
	ecsAPI := ecs.New(sess)

	log.Debug("Connecting to ECR service")
	ecrAPI := ecr.New(sess)

	return &ecsClusterManager{
		ecsAPI: ecsAPI,
		ecrAPI: ecrAPI,
	}, nil
}

// ListInstances get the instances for a specific cluster
func (ecsMgr *ecsClusterManager) ListInstances(clusterName string) ([]*ecs.ContainerInstance, error) {
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

	return describeOut.ContainerInstances, nil
}

func (ecsMgr *ecsClusterManager) AuthenticateRepository(repoURL string) (string, error) {
	ecrAPI := ecsMgr.ecrAPI

	params := &ecr.GetAuthorizationTokenInput{}

	log.Debug("Authenticating to ECR repo")

	resp, err := ecrAPI.GetAuthorizationToken(params)
	if err != nil {
		return Empty, err
	}

	for _, authData := range resp.AuthorizationData {
		if strings.HasPrefix(fmt.Sprintf("https://%s", repoURL), aws.StringValue(authData.ProxyEndpoint)) {
			return aws.StringValue(authData.AuthorizationToken), nil
		}
	}

	return Empty, fmt.Errorf("unable to find token for repo url:%s", repoURL)
}
