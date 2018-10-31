package workflows

import (
	"fmt"

	"github.com/stelligent/mu/common"
)

type purgeWorkflow struct {
	context *common.Context
}

// NewPurge create a new workflow for purging mu resources
func NewPurge(ctx *common.Context) Executor {
	workflow := new(purgeWorkflow)
	workflow.context = ctx

	iamCommonStackName := fmt.Sprintf("%s-iam-common", ctx.Config.Namespace)

	return newPipelineExecutor(
		ctx.RolesetManager.UpsertCommonRoleset,
		workflow.newStackStream(common.StackTypeProduct).foreach(workflow.terminateProduct, workflow.deleteStack),
		workflow.newStackStream(common.StackTypePortfolio).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypePipeline).foreach(workflow.terminatePipeline),
		workflow.newStackStream(common.StackTypeEnv).foreach(workflow.terminateEnvironment),
		workflow.newStackStream(common.StackTypeSchedule).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeService).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeDatabase).foreach(workflow.terminateDatabase),
		workflow.newStackStream(common.StackTypeLoadBalancer).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeEnv).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeVpc).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeTarget).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeRepo).foreach(workflow.cleanupRepo, workflow.deleteStack),
		workflow.newStackStream(common.StackTypeApp).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeProduct).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypePortfolio).foreach(workflow.deleteStack),
		workflow.newStackStream(common.StackTypeBucket).foreach(workflow.cleanupBucket, workflow.deleteStack),
		workflow.newStackStream(common.StackTypeIam).filter(excludeStackName(iamCommonStackName)).foreach(workflow.deleteStack),
		workflow.terminateCommonRoleset(),
	)
}

func (workflow *purgeWorkflow) terminateDatabase(stack *common.Stack) Executor {
	return NewDatabaseTerminator(workflow.context, stack.Tags["service"], stack.Tags["environment"])
}
func (workflow *purgeWorkflow) terminatePipeline(stack *common.Stack) Executor {
	return NewPipelineTerminator(workflow.context, stack.Tags["service"])
}
func (workflow *purgeWorkflow) terminateEnvironment(stack *common.Stack) Executor {
	return NewEnvironmentsTerminator(workflow.context, []string{stack.Tags["environment"]})
}
func (workflow *purgeWorkflow) terminateCommonRoleset() Executor {
	return func() error {
		workflow.context.StackManager.AllowDataLoss(true)
		return workflow.context.RolesetManager.DeleteCommonRoleset()
	}
}
func excludeStackName(stackName string) stackFilter {
	return func(stack *common.Stack) bool {
		return stack.Name != stackName
	}
}

func (workflow *purgeWorkflow) deleteStack(stack *common.Stack) Executor {
	stackName := stack.Name
	return func() error {
		err := workflow.context.StackManager.DeleteStack(stackName)
		if err != nil {
			return err
		}
		status := workflow.context.StackManager.AwaitFinalStatus(stackName)
		if status != nil && !workflow.context.Config.DryRun {
			return fmt.Errorf("Unable to delete stack '%s'", stackName)
		}
		return nil
	}
}

func (workflow *purgeWorkflow) cleanupBucket(stack *common.Stack) Executor {
	bucketName := stack.Outputs["Bucket"]
	return func() error {
		return workflow.context.ArtifactManager.EmptyBucket(bucketName)
	}
}

func (workflow *purgeWorkflow) cleanupRepo(stack *common.Stack) Executor {
	repoName := stack.Parameters["RepoName"]
	return func() error {
		return workflow.context.ClusterManager.DeleteRepository(repoName)
	}
}

type stackStream struct {
	namespace   string
	stackType   common.StackType
	stackLister common.StackLister
	filters     []stackFilter
}

func (workflow *purgeWorkflow) newStackStream(stackType common.StackType) *stackStream {
	return &stackStream{
		namespace:   workflow.context.Config.Namespace,
		stackType:   stackType,
		stackLister: workflow.context.StackManager,
		filters:     make([]stackFilter, 0),
	}
}

type stackExecutor func(stack *common.Stack) Executor
type stackFilter func(stack *common.Stack) bool

func (stream *stackStream) filter(filter stackFilter) *stackStream {
	stream.filters = append(stream.filters, filter)
	return stream
}

// Check if the filters for the stream to see if the stack should be included
func (stream *stackStream) included(stack *common.Stack) bool {
	for _, filter := range stream.filters {
		if !filter(stack) {
			return false
		}
	}
	return true
}

// Create a pipeline executor consiting of the resolved stackExecutors for a given stack
func applyStackExecutors(stack *common.Stack, stackExecutors ...stackExecutor) Executor {
	executors := make([]Executor, 0)
	for _, stackExecutor := range stackExecutors {
		executor := stackExecutor(stack)
		if executor != nil {
			executors = append(executors, executor)
		}
	}
	return newPipelineExecutor(executors...)
}

// Create an executor that can iterate over all stacks and run executors against the stacks
func (stream *stackStream) foreach(stackExecutors ...stackExecutor) Executor {
	return func() error {
		log.Noticef("Purging '%s' stacks in namespace '%s'", stream.stackType, stream.namespace)
		stacks, err := stream.stackLister.ListStacks(stream.stackType, stream.namespace)
		if err != nil {
			return err
		}
		executors := make([]Executor, 0)
		for _, stack := range stacks {
			if !stream.included(stack) {
				log.Debugf("skipping %s", stack.Name)
				continue
			}
			log.Debugf("adding %s", stack.Name)

			executors = append(executors, applyStackExecutors(stack, stackExecutors...))
		}
		executor := newParallelExecutor(executors...)
		return executor()
	}
}
