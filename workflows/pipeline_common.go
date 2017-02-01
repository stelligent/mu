package workflows

import (
	"github.com/fatih/color"
	"github.com/stelligent/mu/common"
)

type pipelineWorkflow struct {
	serviceName    string
	pipelineConfig *common.Pipeline
}

func colorizeActionStatus(actionStatus string) string {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	var color func(a ...interface{}) string
	if actionStatus == "Succeeded" {
		color = green
	} else if actionStatus == "Failed" {
		color = red
	} else {
		color = blue
	}
	return color(actionStatus)
}
