package workflows

import (
	"github.com/fatih/color"
	"github.com/stelligent/mu/common"
)

type pipelineWorkflow struct {
	serviceName    string
	pipelineConfig *common.Pipeline
	codeRevision   string
	repoName       string
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

// Find the service in config
func (workflow *pipelineWorkflow) serviceFinder(serviceName string, ctx *common.Context) Executor {

	return func() error {
		// Repo Name
		if serviceName != "" {
			workflow.serviceName = serviceName
		} else if ctx.Config.Service.Name == "" {
			workflow.serviceName = ctx.Config.Repo.Name
		} else {
			workflow.serviceName = ctx.Config.Service.Name
		}

		workflow.pipelineConfig = &ctx.Config.Service.Pipeline
		workflow.codeRevision = ctx.Config.Repo.Revision

		repoName := ctx.Config.Repo.Slug
		if workflow.pipelineConfig.Source.Repo == "" {
			workflow.pipelineConfig.Source.Repo = repoName
			workflow.repoName = repoName
		} else {
			workflow.repoName = repoName
		}

		if workflow.pipelineConfig.Source.Provider == "" {
			if ctx.Config.Repo.Provider == "" {
				workflow.pipelineConfig.Source.Provider = "GitHub"
			} else {
				workflow.pipelineConfig.Source.Provider = ctx.Config.Repo.Provider
			}
		}
		return nil
	}
}
