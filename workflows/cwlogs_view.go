package workflows

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/stelligent/mu/common"
	"io"
	"strings"
	"sync"
)

type logsWorkflow struct {
}

// NewEnvironmentLogViewer create a new workflow for following logs environments
func NewEnvironmentLogViewer(ctx *common.Context, follow bool, environmentName string, writer io.Writer, filter string) Executor {
	workflow := new(logsWorkflow)

	logGroup := common.CreateStackName(common.StackTypeCluster, environmentName)

	return newWorkflow(
		workflow.logsViewer(ctx.LogsManager, writer, filter, follow, logGroup),
	)
}

// NewServiceLogViewer create a new workflow for following logs for services
func NewServiceLogViewer(ctx *common.Context, follow bool, environmentName string, serviceName string, writer io.Writer, filter string) Executor {
	workflow := new(logsWorkflow)

	if serviceName == "" {
		serviceName = ctx.Config.Service.Name
	}

	logGroup := common.CreateStackName(common.StackTypeService, serviceName, environmentName)

	return newWorkflow(
		workflow.logsViewer(ctx.LogsManager, writer, filter, follow, logGroup),
	)
}

// NewPipelineLogViewer create a new workflow for following logs for pipelines
func NewPipelineLogViewer(ctx *common.Context, follow bool, serviceName string, writer io.Writer, filter string) Executor {
	workflow := new(logsWorkflow)

	if serviceName == "" {
		serviceName = ctx.Config.Service.Name
	}

	var jobs = [...]string{"artifact", "image", "deploy-acceptance", "test-acceptance", "deploy-production", "test-production"}
	var logGroups []string

	for _, job := range jobs {
		logGroups = append(logGroups, fmt.Sprintf("/aws/codebuild/mu-pipeline-%s-%s", serviceName, job))
	}

	return newWorkflow(
		workflow.logsViewer(ctx.LogsManager, writer, filter, follow, logGroups...),
	)
}

func (workflow *logsWorkflow) logsViewer(logsViewer common.LogsViewer, writer io.Writer, filter string, follow bool, logGroups ...string) Executor {
	bold := color.New(color.Bold).SprintFunc()

	return func() error {
		cb := func(logStream string, message string, timestamp int64) {
			fmt.Fprintf(writer, "%s [%s]\n", strings.TrimSpace(message), bold(logStream))
		}

		var wg sync.WaitGroup
		wg.Add(len(logGroups))

		var err error

		for _, logGroup := range logGroups {
			lg := logGroup
			go func() {
				defer wg.Done()
				e := logsViewer.ViewLogs(lg, follow, filter, cb)
				if err == nil {
					err = e
				}
			}()
		}

		wg.Wait()

		return err
	}
}
