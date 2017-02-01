package workflows

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewServiceViewer create a new workflow for showing an environment
func NewServiceViewer(ctx *common.Context, serviceName string, writer io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceInput(ctx, serviceName),
		workflow.serviceViewer(ctx.StackManager, ctx.StackManager, ctx.PipelineManager, writer),
	)
}

func (workflow *serviceWorkflow) serviceInput(ctx *common.Context, serviceName string) Executor {
	return func() error {
		// Repo Name
		if serviceName == "" {
			if ctx.Config.Service.Name == "" {
				workflow.serviceName = ctx.Repo.Name
			} else {
				workflow.serviceName = ctx.Config.Service.Name
			}
		} else {
			workflow.serviceName = serviceName
		}
		return nil
	}
}

func (workflow *serviceWorkflow) serviceViewer(stackLister common.StackLister, stackGetter common.StackGetter, pipelineStateLister common.PipelineStateLister, writer io.Writer) Executor {
	bold := color.New(color.Bold).SprintFunc()
	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeService)
		if err != nil {
			return err
		}

		pipelineStackName := common.CreateStackName(common.StackTypePipeline, workflow.serviceName)
		pipelineStack, err := stackGetter.GetStack(pipelineStackName)
		if err == nil {
			fmt.Fprint(writer, "\n")
			fmt.Fprintf(writer, "%s:\t%s\n", bold("Pipeline URL"), pipelineStack.Outputs["CodePipelineUrl"])

			states, err := pipelineStateLister.ListState(pipelineStack.Outputs["PipelineName"])
			if err != nil {
				return err
			}

			stateTable := buildPipelineStateTable(writer, states)
			stateTable.Render()
			fmt.Fprint(writer, "\n")

		} else {
			fmt.Fprint(writer, "\n")
			fmt.Fprintf(writer, "%s:\t%s\n", bold("Pipeline URL"), "N/A")
		}

		fmt.Fprintf(writer, "%s:\n", bold("Deployments"))

		table := buildEnvTable(writer, stacks, workflow.serviceName)
		table.Render()

		return nil
	}
}

func buildPipelineStateTable(writer io.Writer, stages []*codepipeline.StageState) *tablewriter.Table {
	bold := color.New(color.Bold).SprintFunc()
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Stage", "Action", "Revision", "Status", "Last Update"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, stage := range stages {
		for _, action := range stage.ActionStates {
			revision := "-"
			if action.CurrentRevision != nil {
				revision = aws.StringValue(action.CurrentRevision.RevisionId)
			}
			status := "-"
			message := ""
			lastUpdate := "-"
			if action.LatestExecution != nil {
				lastUpdate = aws.TimeValue(action.LatestExecution.LastStatusChange).Local().Format("2006-01-02 15:04:05")
				status = aws.StringValue(action.LatestExecution.Status)
				if action.LatestExecution.ErrorDetails != nil {
					message = aws.StringValue(action.LatestExecution.ErrorDetails.Message)
				}
			}
			table.Append([]string{
				bold(aws.StringValue(stage.StageName)),
				aws.StringValue(action.ActionName),
				revision,
				fmt.Sprintf("%s %s", colorizeActionStatus(status), message),
				lastUpdate,
			})
		}

	}
	return table
}

func buildEnvTable(writer io.Writer, stacks []*common.Stack, serviceName string) *tablewriter.Table {
	bold := color.New(color.Bold).SprintFunc()
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Environment", "Stack", "Image", "Status", "Last Update", "Mu Version"})
	table.SetBorder(true)
	table.SetAutoWrapText(false)

	for _, stack := range stacks {
		if stack.Tags["service"] != serviceName {
			continue
		}

		table.Append([]string{
			bold(stack.Tags["environment"]),
			stack.Name,
			stack.Parameters["ImageUrl"],
			fmt.Sprintf("%s %s", colorizeStackStatus(stack.Status), stack.StatusReason),
			stack.LastUpdateTime.Local().Format("2006-01-02 15:04:05"),
			stack.Tags["version"],
		})

	}
	return table
}
