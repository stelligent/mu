package workflows

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewServiceViewer create a new workflow for showing an environment
func NewServiceViewer(ctx *common.Context, serviceName string, writer io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
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
			fmt.Fprint(writer, common.NewLine)
			fmt.Fprintf(writer, common.SvcPipelineFormat, common.Bold(common.SvcPipelineURLLabel), pipelineStack.Outputs[common.SvcCodePipelineURLKey])

			states, err := pipelineStateLister.ListState(pipelineStack.Outputs[common.SvcCodePipelineNameKey])
			if err != nil {
				return err
			}

			stateTable := buildPipelineStateTable(writer, states)
			stateTable.Render()
			fmt.Fprint(writer, common.NewLine)

		} else {
			fmt.Fprint(writer, common.NewLine)
			fmt.Fprintf(writer, common.SvcPipelineFormat, common.Bold(common.SvcPipelineURLLabel), common.NA)
		}

		fmt.Fprintf(writer, common.SvcDeploymentsFormat, common.Bold(common.SvcDeploymentsLabel))

		table := buildEnvTable(writer, stacks, workflow.serviceName)
		table.Render()

		viewTasks(taskManager, writer, stacks, workflow.serviceName)

		return nil
	}
}

func buildPipelineStateTable(writer io.Writer, stages []*codepipeline.StageState) *tablewriter.Table {
	table := common.CreateTableSection(writer, common.SvcPipelineTableHeader)

	for _, stage := range stages {
		for _, action := range stage.ActionStates {
			revision := common.LineChar
			if action.CurrentRevision != nil {
				revision = aws.StringValue(action.CurrentRevision.RevisionId)
			}
			status := common.LineChar
			message := common.Empty
			lastUpdate := common.LineChar
			if action.LatestExecution != nil {
				lastUpdate = aws.TimeValue(action.LatestExecution.LastStatusChange).Local().Format(common.LastUpdateTime)
				status = aws.StringValue(action.LatestExecution.Status)
				if action.LatestExecution.ErrorDetails != nil {
					message = aws.StringValue(action.LatestExecution.ErrorDetails.Message)
				}
			}
			table.Append([]string{
				common.Bold(aws.StringValue(stage.StageName)),
				aws.StringValue(action.ActionName),
				revision,
				fmt.Sprintf(common.KeyValueFormat, colorizeActionStatus(status), message),
				lastUpdate,
			})
		}

	}
	return table
}

func buildEnvTable(writer io.Writer, stacks []*common.Stack, serviceName string) *tablewriter.Table {
	table := common.CreateTableSection(writer, common.SvcEnvironmentTableHeader)

	for _, stack := range stacks {
		if stack.Tags[common.SvcCmd] != serviceName {
			continue
		}

		table.Append([]string{
			common.Bold(stack.Tags[common.EnvCmd]),
			stack.Name,
			stack.Parameters[common.SvcImageURLKey],
			fmt.Sprintf(common.KeyValueFormat, colorizeStackStatus(stack.Status), stack.StatusReason),
			stack.LastUpdateTime.Local().Format(common.LastUpdateTime),
			stack.Tags[common.SvcVersionKey],
		})
	}
	return table
}

func viewTasks(taskManager common.TaskManager, writer io.Writer, stacks []*common.Stack, serviceName string) error {
	for _, stack := range stacks {
		if stack.Tags[common.SvcCmd] != serviceName && len(serviceName) != common.Zero {
			continue
		}
		if len(serviceName) == common.Zero {
			serviceName = stack.Tags[common.SvcCmd]
		}
		tasks, err := taskManager.ListTasks(stack.Tags[common.EnvCmd], serviceName)
		if err != nil {
			return err
		}

		fmt.Fprintf(writer, common.SvcContainersFormat, common.Bold(common.SvcContainersLabel), common.Bold(serviceName))
		containersTable := buildTaskTable(tasks, writer)
		containersTable.Render()
	}

	return nil
}

func buildTaskTable(tasks []common.Task, writer io.Writer) *tablewriter.Table {
	table := common.CreateTableSection(writer, common.SvcTaskContainerHeader)
	for _, task := range tasks {
		for _, container := range task.Containers {
			table.Append([]string{
				common.Bold(task.Name),
				container.Name,
				container.Instance,
				container.PrivateIP,
			})
		}
	}
	return table
}
