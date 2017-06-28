package workflows

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"io"
)

// NewEnvironmentLister create a new workflow for listing environments
func NewEnvironmentLister(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(environmentWorkflow)

	return newPipelineExecutor(
		workflow.environmentLister(ctx.StackManager, writer),
	)
}

func (workflow *environmentWorkflow) environmentLister(stackLister common.StackLister, writer io.Writer) Executor {

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeCluster)
		if err != nil {
			return err
		}

		table := CreateTableSection(writer, EnvironmentShowHeader)

		for _, stack := range stacks {
			table.Append([]string{
				Bold(stack.Tags[EnvTagKey]),
				stack.Name,
				fmt.Sprintf(KeyValueFormat, colorizeStackStatus(stack.Status), stack.StatusReason),
				stack.LastUpdateTime.Local().Format(LastUpdateTime),
			})
		}

		table.Render()

		return nil
	}
}
