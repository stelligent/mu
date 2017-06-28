package workflows

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewServiceViewer create a new workflow for showing an environment
func NewServiceViewer(ctx *common.Context, serviceName string, writer io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newPipelineExecutor(
		workflow.serviceInput(ctx, serviceName),
		workflow.serviceViewer(ctx.StackManager, ctx.StackManager, ctx.PipelineManager, ctx.TaskManager, ctx.Config, writer),
	)
}

func (workflow *serviceWorkflow) serviceViewer(stackLister common.StackLister, stackGetter common.StackGetter, pipelineStateLister common.PipelineStateLister, taskManager common.TaskManager, config common.Config, writer io.Writer) Executor {

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeService)
		if err != nil {
			return err
		}

		pipelineStackName := common.CreateStackName(common.StackTypePipeline, workflow.serviceName)
		pipelineStack, err := stackGetter.GetStack(pipelineStackName)
		if err == nil {
			fmt.Fprint(writer, NewLine)
			fmt.Fprintf(writer, SvcPipelineFormat, Bold(SvcPipelineURLLabel), pipelineStack.Outputs[SvcCodePipelineURLKey])

			states, err := pipelineStateLister.ListState(pipelineStack.Outputs[SvcCodePipelineNameKey])
			if err != nil {
				return err
			}

			stateTable := buildPipelineStateTable(writer, states)
			stateTable.Render()
			fmt.Fprint(writer, NewLine)

		} else {
			fmt.Fprint(writer, NewLine)
			fmt.Fprintf(writer, SvcPipelineFormat, Bold(SvcPipelineURLLabel), NA)
		}

		fmt.Fprintf(writer, SvcDeploymentsFormat, Bold(SvcDeploymentsLabel))

		table := buildEnvTable(writer, stacks, workflow.serviceName)
		table.Render()

		viewTasks(taskManager, writer, stacks, workflow.serviceName)

		return nil
	}
}

func buildPipelineStateTable(writer io.Writer, stages []common.PipelineStageState) *tablewriter.Table {
	table := CreateTableSection(writer, SvcPipelineTableHeader)

	for _, stage := range stages {
		for _, action := range stage.ActionStates {
			revision := LineChar
			if action.CurrentRevision != nil {
				revision = common.StringValue(action.CurrentRevision.RevisionId)
			}
			status := LineChar
			message := common.Empty
			lastUpdate := LineChar
			if action.LatestExecution != nil {
				lastUpdate = common.TimeValue(action.LatestExecution.LastStatusChange).Local().Format(LastUpdateTime)
				status = common.StringValue(action.LatestExecution.Status)
				if action.LatestExecution.ErrorDetails != nil {
					message = common.StringValue(action.LatestExecution.ErrorDetails.Message)
				}
			}
			table.Append([]string{
				Bold(common.StringValue(stage.StageName)),
				common.StringValue(action.ActionName),
				revision,
				fmt.Sprintf(KeyValueFormat, colorizeActionStatus(status), message),
				lastUpdate,
			})
		}

	}
	return table
}

func buildEnvTable(writer io.Writer, stacks []*common.Stack, serviceName string) *tablewriter.Table {
	table := CreateTableSection(writer, SvcEnvironmentTableHeader)

	for _, stack := range stacks {
		if stack.Tags[SvcTagKey] != serviceName {
			continue
		}

		table.Append([]string{
			Bold(stack.Tags[EnvTagKey]),
			stack.Name,
			simplifyRepoURL(stack.Parameters[SvcImageURLKey]),
			fmt.Sprintf(KeyValueFormat, colorizeStackStatus(stack.Status), stack.StatusReason),
			stack.LastUpdateTime.Local().Format(LastUpdateTime),
		})
	}
	return table
}

func viewTasks(taskManager common.TaskManager, writer io.Writer, stacks []*common.Stack, serviceName string) error {
	containersTable := CreateTableSection(writer, SvcTaskContainerHeader)
	for _, stack := range stacks {
		if stack.Tags[SvcTagKey] != serviceName && len(serviceName) != Zero {
			continue
		}
		if len(serviceName) == Zero {
			serviceName = stack.Tags[SvcTagKey]
		}
		tasks, err := taskManager.ListTasks(stack.Tags[EnvTagKey], serviceName)
		if err != nil {
			return err
		}

		for _, task := range tasks {
			for _, container := range task.Containers {
				containersTable.Append([]string{
					stack.Tags[EnvTagKey],
					container.Name,
					Bold(task.Name),
					container.Instance,
				})
			}
		}

	}

	fmt.Fprintf(writer, SvcContainersFormat, Bold(SvcContainersLabel), Bold(serviceName))
	containersTable.Render()

	return nil
}
