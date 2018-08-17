package workflows

import (
	"fmt"
	"io"

	"github.com/stelligent/mu/common"
)

// NewDatabaseLister create a new workflow for listing databases
func NewDatabaseLister(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(databaseWorkflow)

	return newPipelineExecutor(
		workflow.databaseLister(ctx.Config.Namespace, ctx.StackManager, writer),
	)
}

func (workflow *databaseWorkflow) databaseLister(namespace string, stackLister common.StackLister, writer io.Writer) Executor {

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeDatabase, namespace)
		if err != nil {
			return err
		}

		table := CreateTableSection(writer, PipeLineServiceHeader)

		for _, stack := range stacks {

			table.Append([]string{
				Bold(stack.Tags[SvcTagKey]),
				stack.Name,
				fmt.Sprintf(KeyValueFormat, colorizeStackStatus(stack.Status), stack.StatusReason),
				stack.LastUpdateTime.Local().Format(LastUpdateTime),
			})

		}

		table.Render()

		return nil
	}
}
