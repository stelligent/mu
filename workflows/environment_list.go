package workflows

import (
	"fmt"
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

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeCluster)
		if err != nil {
			return err
		}

		table := common.CreateTableSection(writer, common.EnvironmentShowHeader)

		for _, stack := range stacks {
			table.Append([]string{
				common.Bold(stack.Tags[common.EnvCmd]),
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
