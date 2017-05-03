package workflows

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewDatabaseLister create a new workflow for listing databases
func NewDatabaseLister(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(databaseWorkflow)

	return newWorkflow(
		workflow.databaseLister(ctx.StackManager, writer),
	)
}

func (workflow *databaseWorkflow) databaseLister(stackLister common.StackLister, writer io.Writer) Executor {
	bold := color.New(color.Bold).SprintFunc()

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeDatabase)

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
