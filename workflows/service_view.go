package workflows

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/stelligent/mu/common"
	"io"
)

// NewServiceViewer create a new workflow for showing an environment
func NewServiceViewer(ctx *common.Context, serviceName string, writer io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newWorkflow(
		workflow.serviceInput(ctx, serviceName),
		workflow.serviceViewer(ctx.StackManager, ctx.ClusterManager, writer),
	)
}

func (workflow *serviceWorkflow) serviceInput(ctx *common.Context, serviceName string) Executor {
	return func() error {
		// Repo Name
		if serviceName == "" {
			if ctx.Config.Service.Name == "" {
				workflow.serviceName = ctx.Repo.Name
			} else {
				workflow.serviceName = ctx.Config.Service.Name
			}
		} else {
			workflow.serviceName = serviceName
		}
		return nil
	}
}

func (workflow *serviceWorkflow) serviceViewer(stackLister common.StackLister, instanceLister common.ClusterInstanceLister, writer io.Writer) Executor {
	bold := color.New(color.Bold).SprintFunc()
	return func() error {
		stacks, err := stackLister.ListStacks(common.StackTypeService)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(writer)
		table.SetHeader([]string{"Environment", "Stack", "Image", "Status", "Last Update", "Mu Version"})
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		for _, stack := range stacks {
			if stack.Tags["service"] != workflow.serviceName {
				continue
			}

			table.Append([]string{
				bold(stack.Tags["environment"]),
				stack.Name,
				stack.Parameters["ImageUrl"],
				fmt.Sprintf("%s %s", colorizeStackStatus(stack.Status), stack.StatusReason),
				stack.LastUpdateTime.Local().Format("2006-01-02 15:04:05"),
				stack.Tags["version"],
			})

		}

		table.Render()

		return nil
	}
}
