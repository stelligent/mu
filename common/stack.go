package common

import (
	"fmt"
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
	UpsertStack(stackName string, templateName string, templateData interface{}, parameters map[string]string, tags map[string]string, policy string, roleArn string) error
	SetTerminationProtection(stackName string, enabled bool) error
}

// StackLister for listing stacks
type StackLister interface {
	ListStacks(stackType StackType, namespace string) ([]*Stack, error)
}

// StackGetter for getting stacks
type StackGetter interface {
	GetStack(stackName string) (*Stack, error)
}

// StackDeleter for deleting stacks
type StackDeleter interface {
	DeleteStack(stackName string) error
}

// ImageFinder for finding latest image
type ImageFinder interface {
	FindLatestImageID(owner string, namePattern string) (string, error)
}

// AZCounter for counting availability zones in a region
type AZCounter interface {
	CountAZs() (int, error)
}

// StackManager composite of all stack capabilities
type StackManager interface {
	StackUpserter
	StackWaiter
	StackLister
	StackGetter
	StackDeleter
	ImageFinder
	AZCounter
	AllowDataLoss(allow bool)
}
