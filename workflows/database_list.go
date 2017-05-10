package workflows

import (
	"fmt"
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

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeDatabase)
		if err != nil {
			return err
		}

		table := common.CreateTableSection(writer, common.PipeLineServiceHeader)

		for _, stack := range stacks {

			table.Append([]string{
				common.Bold(stack.Tags[common.SvcCmd]),
				stack.Name,
				fmt.Sprintf(common.KeyValueFormat, colorizeStackStatus(stack.Status), stack.StatusReason),
				stack.LastUpdateTime.Local().Format(common.LastUpdateTime),
				stack.Tags[common.SvcVersionKey],
			})
		}

		table.Render()

		return nil
	}
}
