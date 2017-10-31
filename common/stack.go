package common

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// CreateStackName will create a name for a stack
func CreateStackName(namespace string, stackType StackType, names ...string) string {
	return fmt.Sprintf("%s-%s-%s", namespace, stackType, strings.Join(names, "-"))
}

// GetStackOverrides will get the overrides from the config
func GetStackOverrides(stackName string) []interface{} {
	resp := make([]interface{}, 0)

	for _, stackOverride := range stackOverrides {
		if stackOverride.stackNameMatcher.MatchString(stackName) {
			resp = append(resp, stackOverride)
		}
	}

	return resp
}

type _StackOverride struct {
	stackNameMatcher *regexp.Regexp
	template         interface{}
}

var stackOverrides = make([]_StackOverride, 0)

func registerStackOverride(stackNamePattern string, template interface{}) {
	stackOverrides = append(stackOverrides, _StackOverride{
		regexp.MustCompile(stackNamePattern),
		template,
	})
}

// StackWaiter for waiting on stack status to be final
type StackWaiter interface {
	AwaitFinalStatus(stackName string) *Stack
}

// StackUpserter for applying changes to a stack
type StackUpserter interface {
	UpsertStack(stackName string, templateBodyReader io.Reader, parameters map[string]string, tags map[string]string, roleArn string) error
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
