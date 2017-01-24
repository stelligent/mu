package workflows

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewEnvironmentLister create a new workflow for listing environments
func NewEnvironmentLister(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	return newWorkflow(
		workflow.environmentLister(ctx.StackManager, writer),
	)
}

func (workflow *environmentWorkflow) environmentLister(stackLister common.StackLister, writer io.Writer) Executor {
	bold := color.New(color.Bold).SprintFunc()

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeCluster)

		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(writer)
		table.SetHeader([]string{"Environment", "Stack", "Status", "Last Update", "Mu Version"})
		table.SetBorder(false)

		for _, stack := range stacks {

			table.Append([]string{
				bold(stack.Tags["environment"]),
				stack.Name,
				fmt.Sprintf("%s %s", colorizeStackStatus(stack.Status), stack.StatusReason),
				stack.LastUpdateTime.String(),
				stack.Tags["version"],
			})

		}

		table.Render()

		return nil
	}
}
