package workflows

import (
	"github.com/fatih/color"
	"github.com/stelligent/mu/common"
	"strings"
)

type environmentWorkflow struct {
	environment  *common.Environment
	codeRevision string
	repoName     string
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
