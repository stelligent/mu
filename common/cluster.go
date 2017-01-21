package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/ecr"
	"strings"
	"fmt"
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

func newClusterManager(region string) (ClusterManager, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	log.Debugf("Connecting to ECS service in region:%s", region)
	ecsAPI := ecs.New(sess, &aws.Config{Region: aws.String(region)})

	log.Debugf("Connecting to ECR service in region:%s", region)
	ecrAPI := ecr.New(sess, &aws.Config{Region: aws.String(region)})

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
	describeOut, err := ecsAPI.DescribeContainerInstances(describeParams)

	return describeOut.ContainerInstances, nil
}

func (ecsMgr *ecsClusterManager) AuthenticateRepository(repoURL string) (string, error) {
	ecrAPI := ecsMgr.ecrAPI

	params := &ecr.GetAuthorizationTokenInput{}

	log.Debug("Authenticating to ECR repo")

	resp, err := ecrAPI.GetAuthorizationToken(params)
	if err != nil {
		return "", err
	}

	for _, authData := range resp.AuthorizationData {
		if strings.HasPrefix(fmt.Sprintf("https://%s",repoURL), aws.StringValue(authData.ProxyEndpoint)) {
			return aws.StringValue(authData.AuthorizationToken), nil
		}
	}

	return "", fmt.Errorf("unable to find token for repo url:%s",repoURL)
}
