package workflows

import (
	"github.com/stelligent/mu/common"
)

// NewCatalogTerminator create a new workflow for terminating a catalog
func NewCatalogTerminator(ctx *common.Context) Executor {

	workflow := new(purgeWorkflow)
	workflow.context = ctx

	return newPipelineExecutor(
		workflow.newStackStream(common.StackTypeProduct).foreach(workflow.terminateProduct, workflow.deleteStack),
		workflow.newStackStream(common.StackTypePortfolio).foreach(workflow.deleteStack),
	)

}

func (workflow *purgeWorkflow) terminateProduct(stack *common.Stack) Executor {
	return func() error {
		return workflow.context.CatalogManager.TerminateProvisionedProducts(stack.Outputs["ProductId"])
	}
}
