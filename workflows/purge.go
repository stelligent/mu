package workflows

import (
	"bufio"
	"fmt"
	"github.com/stelligent/mu/common"
	"io"
	"os"
	"strings"
)

type purgeWorkflow struct{}

type stackTerminateWorkflow struct {
	Stack *common.Stack
}

// NewPurge create a new workflow for purging mu resources
func NewPurge(ctx *common.Context, suppressConfirmation bool, writer io.Writer) Executor {
	workflow := new(purgeWorkflow)

	return newPipelineExecutor(
		workflow.purgeWorker(ctx, suppressConfirmation, ctx.StackManager, writer),
	)
}

func filterStacksByStatus(stacks []*common.Stack, statuses []string) []*common.Stack {
	var ret []*common.Stack
	for _, stack := range stacks {
		found := false
		for _, status := range statuses {
			if stack.Status == status {
				found = true
			}
		}
		if !found {
			ret = append(ret, stack)
		}
	}
	return ret
}

func filterStacksByType(stacks []*common.Stack, stackType common.StackType) []*common.Stack {
	var ret []*common.Stack
	for _, stack := range stacks {
		if stack.Tags["type"] == string(stackType) {
			ret = append(ret, stack)
		}
	}
	return ret
}

func (workflow *stackTerminateWorkflow) stackTerminator(ctx *common.Context, stackDeleter common.StackDeleter, stackLister common.StackLister, stackWaiter common.StackWaiter) Executor {
	return func() error {
		// delete the stack object
		err := stackDeleter.DeleteStack(workflow.Stack.Name)
		if err != nil {
			log.Errorf("DeleteStack %s %v", workflow.Stack.Name, err)
		}
		// wait for the result
		svcStack := stackWaiter.AwaitFinalStatus(workflow.Stack.Name)
		if svcStack != nil && !strings.HasSuffix(svcStack.Status, "_COMPLETE") {
			log.Errorf("Ended in failed status %s %s", svcStack.Status, svcStack.StatusReason)
		}
		return nil
	}
}

func (workflow *purgeWorkflow) purgeWorker(ctx *common.Context, suppressConfirmation bool, stackLister common.StackLister, writer io.Writer) Executor {
	return func() error {
		// gather all the stackNames for each type (in parallel)
		stacks, err := stackLister.ListAllStacks()
		if err != nil {
			log.Warning("couldn't list stacks (all)")
		}
		// ignore those in ROLLBACK_COMPLETED status
		// stacks = filterStacksByStatus(stacks, []string{cloudformation.StackStatusRollbackComplete})

		table := CreateTableSection(writer, PurgeHeader)
		stackCount := 0
		for _, stack := range stacks {
			stackType, ok := stack.Tags["type"]
			if ok {
				table.Append([]string{
					Bold(stackType),
					stack.Name,
					fmt.Sprintf(KeyValueFormat, colorizeStackStatus(stack.Status), stack.StatusReason),
					stack.LastUpdateTime.Local().Format(LastUpdateTime),
				})
				stackCount++
			}
		}
		table.Render()

		// create a grand master list of all the things we're going to delete
		var executors []Executor

		// scheduled tasks are attached to services, so they must be deleted first.
		for _, scheduleStack := range filterStacksByType(stacks, common.StackTypeSchedule) {
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = scheduleStack
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		// add the services we're going to terminate
		svcWorkflow := new(serviceWorkflow)
		for _, stack := range filterStacksByType(stacks, common.StackTypeService) {
			executors = append(executors, svcWorkflow.serviceInput(ctx, stack.Tags["service"]))
			executors = append(executors, svcWorkflow.serviceUndeployer(ctx.Config.Namespace, stack.Tags["environment"], ctx.StackManager, ctx.StackManager))
		}

		// Add the terminator jobs to the master list for each environment
		envWorkflow := new(environmentWorkflow)
		for _, stack := range filterStacksByType(stacks, common.StackTypeEnv) {
			// Add the terminator jobs to the master list for each environment
			envName := stack.Tags["environment"]
			executors = append(executors, envWorkflow.environmentServiceTerminator(envName, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.RolesetManager))
			executors = append(executors, envWorkflow.environmentDbTerminator(envName, ctx.StackManager, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentEcsTerminator(ctx.Config.Namespace, envName, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentConsulTerminator(ctx.Config.Namespace, envName, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentRolesetTerminator(ctx.RolesetManager, envName))
			executors = append(executors, envWorkflow.environmentElbTerminator(ctx.Config.Namespace, envName, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentVpcTerminator(ctx.Config.Namespace, envName, ctx.StackManager, ctx.StackManager))
		}

		// add the consuls to terminate
		for _, consul := range filterStacksByType(stacks, common.StackTypeConsul) {
			log.Infof("%s %v", consul.Name, consul.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = consul
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		// add the pipelines to terminate
		codePipelineWorkflow := new(pipelineWorkflow)
		for _, codePipeline := range filterStacksByType(stacks, common.StackTypePipeline) {
			executors = append(executors, codePipelineWorkflow.serviceFinder(codePipeline.Tags["service"], ctx))
			executors = append(executors, codePipelineWorkflow.pipelineTerminator(ctx.Config.Namespace, ctx.StackManager, ctx.StackManager))
			executors = append(executors, codePipelineWorkflow.pipelineRolesetTerminator(ctx.RolesetManager))
		}

		// add the buckets to remove
		for _, bucket := range filterStacksByType(stacks, common.StackTypeBucket) {
			log.Infof("%s %v", bucket.Name, bucket.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = bucket
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		// add the ecr repos to remove
		for _, repo := range filterStacksByType(stacks, common.StackTypeRepo) {
			log.Infof("%s %v", repo.Name, repo.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = repo
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		// add the vpc to delete
		for _, vpcStack := range filterStacksByType(stacks, common.StackTypeVpc) {
			log.Infof("%s %v", vpcStack.Name, vpcStack.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = vpcStack
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		// add the iam roles to delete
		for _, roleStack := range filterStacksByType(stacks, common.StackTypeIam) {
			log.Infof("%s %v", roleStack.Name, roleStack.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = roleStack
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		log.Infof("total of %d stacks to purge", stackCount)

		// prompt the user if -y not present on command-line
		if suppressConfirmation == false {
			phrase := "Yes, I want to purge everything."
			fmt.Printf("Are you sure you want to purge the above resources? ")
			fmt.Printf("(use '%s' to confirm): ", phrase)
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			confirmation := scanner.Text()
			if confirmation != phrase {
				log.Errorf("Aborting at user request")
				os.Exit(1)
			}
		}

		// newPipelineExecutorNoStop is just like newPipelineExecutor, except that it doesn't stop on error
		executor := newPipelineExecutorNoStop(executors...)

		// run everything we've collected
		executor()
		return nil
	}
}
