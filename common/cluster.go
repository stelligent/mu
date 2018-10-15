package common

import "github.com/aws/aws-sdk-go/service/ecs"

// ContainerInstance represents the ECS container instance
type ContainerInstance *ecs.ContainerInstance

// ClusterInstanceLister for getting cluster instances
type ClusterInstanceLister interface {
	ListInstances(clusterName string) ([]ContainerInstance, error)
}

// RepositoryAuthenticator auths for a repo
type RepositoryAuthenticator interface {
	AuthenticateRepository(repoURL string) (string, error)
}

// RepositoryDeleter deletes a repo
type RepositoryDeleter interface {
	DeleteRepository(repoName string) error
}

// ClusterManager composite of all cluster capabilities
type ClusterManager interface {
	ClusterInstanceLister
	RepositoryAuthenticator
	RepositoryDeleter
}
