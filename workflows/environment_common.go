package workflows

import (
	"github.com/fatih/color"
	"github.com/stelligent/mu/common"
	"strings"
)

type environmentWorkflow struct {
	environment           *common.Environment
	codeRevision          string
	repoName              string
	cloudFormationRoleArn string
}

func colorizeStackStatus(stackStatus string) string {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	var color func(a ...interface{}) string
	if strings.HasSuffix(stackStatus, "_FAILED") {
		color = red
	} else if strings.HasSuffix(stackStatus, "_COMPLETE") {
		color = green
	} else {
		color = blue
	}
	return color(stackStatus)
}

func (workflow *environmentWorkflow) isConsulEnabled() Conditional {
	return func() bool {
		return strings.EqualFold(workflow.environment.Discovery.Provider, "consul")
	}
}

func (workflow *environmentWorkflow) isEcsProvider() Conditional {
	return func() bool {
		return strings.EqualFold(string(workflow.environment.Provider), string(common.EnvProviderEcs))
	}
}

func (workflow *environmentWorkflow) isEc2Provider() Conditional {
	return func() bool {
		return strings.EqualFold(string(workflow.environment.Provider), string(common.EnvProviderEc2))
	}
}
