package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

// ClusterInstanceLister for getting cluster instances
type ClusterInstanceLister interface {
	ListInstances(clusterName string) ([]*ecs.ContainerInstance, error)
}

// ClusterManager composite of all cluster capabilities
type ClusterManager interface {
	ClusterInstanceLister
}

type ecsClusterManager struct {
	ecsAPI ecsiface.ECSAPI
}

func newClusterManager(region string) (ClusterManager, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	log.Debugf("Connecting to ECS service in region:%s", region)
	ecs := ecs.New(sess, &aws.Config{Region: aws.String(region)})
	return &ecsClusterManager{
		ecsAPI: ecs,
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
