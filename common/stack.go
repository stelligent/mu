package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"strings"
)

// CreateStackName will create a name for a stack
func CreateStackName(namespace string, stackType StackType, names ...string) string {
	return fmt.Sprintf("%s-%s-%s", namespace, stackType, strings.Join(names, "-"))
}

// StackWaiter for waiting on stack status to be final
type StackWaiter interface {
	AwaitFinalStatus(stackName string) *Stack
}

// StackUpserter for applying changes to a stack
type StackUpserter interface {
	UpsertStack(stackName string, templateName string, templateData interface{}, parameters map[string]string, tags map[string]string, roleArn string) error
}

// StackLister for listing stacks
type StackLister interface {
	ListAllStacks() ([]*Stack, error)
	ListStacks(stackType StackType) ([]*Stack, error)
	GetResourcesForStack(stack *Stack) ([]*cloudformation.StackResource, error)
}

// StackGetter for getting stacks
type StackGetter interface {
	GetStack(stackName string) (*Stack, error)
}

// StackDeleter for deleting stacks
type StackDeleter interface {
	DeleteStack(stackName string) error
}

// S3StackDeleter for purging files from S3 buckets
type S3StackDeleter interface {
	DeleteS3Bucket(bucketName string) error
	DeleteS3BucketObjects(bucketName string) error
}

// EcrRepoDeleter interface {
type EcrRepoDeleter interface {
	DeleteImagesFromEcrRepo(repoName string) error
}

// RoleDeleter interface {
type RoleDeleter interface {
	DeleteRolesForNamespace(namespace string) error
}

// ImageFinder for finding latest image
type ImageFinder interface {
	FindLatestImageID(namePattern string) (string, error)
}

// StackManager composite of all stack capabilities
type StackManager interface {
	StackUpserter
	StackWaiter
	StackLister
	StackGetter
	StackDeleter
	S3StackDeleter
	EcrRepoDeleter
	RoleDeleter
	ImageFinder
}
