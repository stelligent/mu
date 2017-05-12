package workflows

import (
	"fmt"
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

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypePipeline)
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
