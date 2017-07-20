package common

import (
	"fmt"
	"io"
	"strings"
)

// CreateStackName will create a name for a stack
func CreateStackName(stackType StackType, names ...string) string {
	return fmt.Sprintf("mu-%s-%s", stackType, strings.Join(names, "-"))
}

// GetStackOverrides will get the overrides from the config
func GetStackOverrides(stackName string) interface{} {
	if stackOverrides == nil {
		return nil
	}

	return stackOverrides[stackName]
}

var stackOverrides map[string]interface{}

func registerStackOverrides(overrides map[string]interface{}) {
	stackOverrides = overrides
}

// StackWaiter for waiting on stack status to be final
type StackWaiter interface {
	AwaitFinalStatus(stackName string) *Stack
}

// StackUpserter for applying changes to a stack
type StackUpserter interface {
	UpsertStack(stackName string, templateBodyReader io.Reader, parameters map[string]string, tags map[string]string) error
}

// StackLister for listing stacks
type StackLister interface {
	ListStacks(stackType StackType) ([]*Stack, error)
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
	FindLatestImageID(namePattern string) (string, error)
}

// StackManager composite of all stack capabilities
type StackManager interface {
	StackUpserter
	StackWaiter
	StackLister
	StackGetter
	StackDeleter
	ImageFinder
}
