package workflows

import (
	"fmt"
	"io"

	"github.com/stelligent/mu/common"
)

// NewPipelineLister create a new workflow for listing environments
func NewPipelineLister(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(pipelineWorkflow)

	return newPipelineExecutor(
		workflow.pipelineLister(ctx.Config.Namespace, ctx.StackManager, writer),
	)
}

func (workflow *pipelineWorkflow) pipelineLister(namespace string, stackLister common.StackLister, writer io.Writer) Executor {

	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypePipeline, namespace)
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
