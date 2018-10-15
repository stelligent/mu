package workflows

import (
	"fmt"

	"github.com/stelligent/mu/common"
)

type purgeWorkflow struct{}

// NewPurge create a new workflow for purging mu resources
func NewPurge(ctx *common.Context) Executor {
	workflow := new(purgeWorkflow)

	return newPipelineExecutor(
		workflow.terminatePipelines(ctx),
		workflow.terminateEnvironments(ctx),
		workflow.terminateRepos(ctx),
		workflow.terminateApps(ctx),
		workflow.terminateBuckets(ctx),
		workflow.terminateIAM(ctx),
	)
}

func (workflow *purgeWorkflow) terminateIAM(ctx *common.Context) Executor {
	return func() error {
		namespace := ctx.Config.Namespace
		stacks, err := ctx.StackManager.ListStacks(common.StackTypeIam, namespace)
		if err != nil {
			return err
		}
		executors := make([]Executor, 0)
		for _, stack := range stacks {
			stackName := stack.Name
			if stackName != fmt.Sprintf("%s-iam-common", namespace) {
				executors = append(executors, func() error {
					err = ctx.StackManager.DeleteStack(stackName)
					if err != nil {
						return err
					}
					ctx.StackManager.AwaitFinalStatus(stackName)
					return nil
				})
			}
		}
		iamExecutors := newParallelExecutor(executors...)
		err = iamExecutors()
		if err != nil {
			return err
		}

		ctx.StackManager.AllowDataLoss(true)
		return ctx.RolesetManager.DeleteCommonRoleset()
	}
}

func (workflow *purgeWorkflow) terminatePipelines(ctx *common.Context) Executor {
	namespace := ctx.Config.Namespace
	stacks, err := ctx.StackManager.ListStacks(common.StackTypePipeline, namespace)
	if err != nil {
		return newErrorExecutor(err)
	}
	executors := make([]Executor, 0)
	for _, stack := range stacks {
		executors = append(executors, NewPipelineTerminator(ctx, stack.Tags["service"]))
	}
	return newParallelExecutor(executors...)
}

func (workflow *purgeWorkflow) terminateEnvironments(ctx *common.Context) Executor {
	stackLister := ctx.StackManager
	namespace := ctx.Config.Namespace
	stacks, err := stackLister.ListStacks(common.StackTypeEnv, namespace)
	if err != nil {
		return newErrorExecutor(err)
	}
	envNames := make([]string, 0)
	for _, stack := range stacks {
		envNames = append(envNames, stack.Tags["environment"])
	}
	return NewEnvironmentsTerminator(ctx, envNames)
}

func (workflow *purgeWorkflow) terminateRepos(ctx *common.Context) Executor {
	namespace := ctx.Config.Namespace
	stacks, err := ctx.StackManager.ListStacks(common.StackTypeRepo, namespace)
	if err != nil {
		return newErrorExecutor(err)
	}
	executors := make([]Executor, 0)
	for _, stack := range stacks {
		stackName := stack.Name
		repoName := stack.Parameters["RepoName"]
		executors = append(executors, func() error {
			err := ctx.ClusterManager.DeleteRepository(repoName)
			if err != nil {
				return err
			}
			err = ctx.StackManager.DeleteStack(stackName)
			if err != nil {
				return err
			}
			ctx.StackManager.AwaitFinalStatus(stackName)
			return nil
		})
	}
	return newParallelExecutor(executors...)
}

func (workflow *purgeWorkflow) terminateApps(ctx *common.Context) Executor {
	namespace := ctx.Config.Namespace
	stacks, err := ctx.StackManager.ListStacks(common.StackTypeApp, namespace)
	if err != nil {
		return newErrorExecutor(err)
	}
	executors := make([]Executor, 0)
	for _, stack := range stacks {
		stackName := stack.Name
		executors = append(executors, func() error {
			err = ctx.StackManager.DeleteStack(stackName)
			if err != nil {
				return err
			}
			ctx.StackManager.AwaitFinalStatus(stackName)
			return nil
		})
	}
	return newParallelExecutor(executors...)
}

func (workflow *purgeWorkflow) terminateBuckets(ctx *common.Context) Executor {
	namespace := ctx.Config.Namespace
	stacks, err := ctx.StackManager.ListStacks(common.StackTypeBucket, namespace)
	if err != nil {
		return newErrorExecutor(err)
	}
	executors := make([]Executor, 0)
	for _, stack := range stacks {
		stackName := stack.Name
		bucketName := stack.Outputs["Bucket"]
		executors = append(executors, func() error {
			err := ctx.ArtifactManager.EmptyBucket(bucketName)
			if err != nil {
				return err
			}
			err = ctx.StackManager.DeleteStack(stackName)
			if err != nil {
				return err
			}
			ctx.StackManager.AwaitFinalStatus(stackName)
			return nil
		})
	}
	return newParallelExecutor(executors...)
}
