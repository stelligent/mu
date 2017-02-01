package workflows

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewPipelineLister create a new workflow for listing environments
func NewPipelineLister(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(pipelineWorkflow)

	return newWorkflow(
		workflow.pipelineLister(ctx.StackManager, writer),
	)
}

func (workflow *pipelineWorkflow) pipelineLister(stackLister common.StackLister, writer io.Writer) Executor {
	bold := color.New(color.Bold).SprintFunc()

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypePipeline)

		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(writer)
		table.SetHeader([]string{"Service", "Stack", "Status", "Last Update", "Mu Version"})
		table.SetBorder(true)
		table.SetAutoWrapText(false)

		for _, stack := range stacks {

			table.Append([]string{
				bold(stack.Tags["service"]),
				stack.Name,
				fmt.Sprintf("%s %s", colorizeStackStatus(stack.Status), stack.StatusReason),
				stack.LastUpdateTime.Local().Format("2006-01-02 15:04:05"),
				stack.Tags["version"],
			})
		}

		table.Render()

		return nil
	}
}
