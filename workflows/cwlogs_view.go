package workflows

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/stelligent/mu/common"
)

type logsWorkflow struct {
}

// NewEnvironmentLogViewer create a new workflow for following logs environments
func NewEnvironmentLogViewer(ctx *common.Context, searchDuration time.Duration, follow bool, environmentName string, writer io.Writer, filter string) Executor {
	workflow := new(logsWorkflow)

	logGroup := common.CreateStackName(ctx.Config.Namespace, common.StackTypeEnv, environmentName)

	return newPipelineExecutor(
		workflow.logsViewer(ctx.LogsManager, writer, filter, searchDuration, follow, logGroup),
	)
}

// NewServiceLogViewer create a new workflow for following logs for services
func NewServiceLogViewer(ctx *common.Context, searchDuration time.Duration, follow bool, environmentName string, serviceName string, writer io.Writer, filter string) Executor {
	workflow := new(logsWorkflow)

	logGroup := common.CreateStackName(ctx.Config.Namespace, common.StackTypeService, getServiceName(ctx, serviceName), environmentName)

	return newPipelineExecutor(
		workflow.logsViewer(ctx.LogsManager, writer, filter, searchDuration, follow, logGroup),
	)
}

// NewPipelineLogViewer create a new workflow for following logs for pipelines
func NewPipelineLogViewer(ctx *common.Context, searchDuration time.Duration, follow bool, serviceName string, writer io.Writer, filter string) Executor {
	workflow := new(logsWorkflow)

	var jobs = [...]string{"artifact", "image", "deploy-acceptance", "test-acceptance", "deploy-production", "test-production"}
	var logGroups []string

	for _, job := range jobs {
		logGroups = append(logGroups, fmt.Sprintf("/aws/codebuild/%s-pipeline-%s-%s", ctx.Config.Namespace, getServiceName(ctx, serviceName), job))
	}

	return newPipelineExecutor(
		workflow.logsViewer(ctx.LogsManager, writer, filter, searchDuration, follow, logGroups...),
	)
}

func getServiceName(ctx *common.Context, serviceName string) string {
	if serviceName == "" {
		if ctx.Config.Service.Name != "" {
			serviceName = ctx.Config.Service.Name
		} else if ctx.Config.Repo.Name != "" {
			serviceName = ctx.Config.Repo.Name
		}
	}
	return serviceName
}

func (workflow *logsWorkflow) logsViewer(logsViewer common.LogsViewer, writer io.Writer, filter string, searchDuration time.Duration, follow bool, logGroups ...string) Executor {

	return func() error {
		cb := func(logStream string, message string, timestamp int64) {
			// TODO: unchecked return
			fmt.Fprintf(writer, "[%s] %s\n", Bold(logStream), strings.TrimSpace(message))
		}

		var wg sync.WaitGroup
		wg.Add(len(logGroups))

		var err error

		for _, logGroup := range logGroups {
			lg := logGroup
			go func() {
				defer wg.Done()
				e := logsViewer.ViewLogs(lg, searchDuration, follow, filter, cb)
				if err == nil {
					err = e
				}
			}()
		}

		wg.Wait()

		return err
	}
}
