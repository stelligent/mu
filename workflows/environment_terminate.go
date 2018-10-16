package workflows

import (
	"fmt"
	"strings"

	"github.com/stelligent/mu/common"
)

// NewEnvironmentsTerminator create a new workflow for terminating an environment
func NewEnvironmentsTerminator(ctx *common.Context, environmentNames []string) Executor {
	envWorkflows := make([]Executor, len(environmentNames))
	for i, environmentName := range environmentNames {
		envWorkflows[i] = newEnvironmentTerminator(ctx, environmentName)
	}
	return newParallelExecutor(envWorkflows...)
}

func newEnvironmentTerminator(ctx *common.Context, environmentName string) Executor {

	workflow := new(environmentWorkflow)

	return newPipelineExecutor(
		workflow.environmentLoader(ctx.Config.Namespace, environmentName, ctx.StackManager, &environmentView{}),
		workflow.environmentServiceTerminator(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.RolesetManager),
		workflow.environmentDbTerminator(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager, ctx.StackManager),
		newConditionalExecutor(workflow.isKubernetesProvider(),
			newPipelineExecutor(
				workflow.connectKubernetes(ctx.Config.Namespace, ctx.KubernetesResourceManagerProvider),
				workflow.environmentKubernetesIngressTerminator(environmentName),
				workflow.environmentEcsTerminator(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager),
				workflow.environmentRolesetTerminator(ctx.RolesetManager, environmentName),
			),
			newPipelineExecutor(
				workflow.environmentEcsTerminator(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager),
				workflow.environmentRolesetTerminator(ctx.RolesetManager, environmentName),
				workflow.environmentElbTerminator(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager),
			),
		),
		workflow.environmentVpcTerminator(ctx.Config.Namespace, environmentName, ctx.StackManager, ctx.StackManager),
	)
}

func (workflow *environmentWorkflow) environmentServiceTerminator(namespace string, environmentName string, stackLister common.StackLister, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter, rolesetDeleter common.RolesetDeleter) Executor {
	return func() error {
		log.Noticef("Terminating Services for environment '%s' ...", environmentName)
		stacks, err := stackLister.ListStacks(common.StackTypeService, namespace)
		if err != nil {
			return err
		}

		executors := make([]Executor, 0)
		for _, stack := range stacks {
			if stack.Tags["environment"] != environmentName {
				continue
			}
			err := stackDeleter.DeleteStack(stack.Name)
			if err != nil {
				return err
			}

			serviceName := stack.Tags["service"]
			stackName := stack.Name
			executors = append(executors, func() error {
				log.Infof("   Undeploying service '%s' from environment '%s'", serviceName, environmentName)
				stackWaiter.AwaitFinalStatus(stackName)
				return rolesetDeleter.DeleteServiceRoleset(environmentName, serviceName)
			})
		}

		return newParallelExecutor(executors...)()
	}
}
func (workflow *environmentWorkflow) environmentDbTerminator(namespace string, environmentName string, stackLister common.StackLister, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Terminating Databases for environment '%s' ...", environmentName)
		stacks, err := stackLister.ListStacks(common.StackTypeDatabase, namespace)
		if err != nil {
			return err
		}
		for _, stack := range stacks {
			if stack.Tags["environment"] != environmentName {
				continue
			}
			err := stackDeleter.DeleteStack(stack.Name)
			if err != nil {
				return err
			}
		}
		for _, stack := range stacks {
			if stack.Tags["environment"] != environmentName {
				continue
			}
			log.Infof("   Terminating database for service '%s' from environment '%s'", stack.Tags["service"], environmentName)
			stackWaiter.AwaitFinalStatus(stack.Name)
		}

		return nil
	}
}
func (workflow *environmentWorkflow) environmentEcsTerminator(namespace string, environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Terminating environment '%s' ...", environmentName)
		envStackName := common.CreateStackName(namespace, common.StackTypeEnv, environmentName)
		err := stackDeleter.DeleteStack(envStackName)
		if err != nil {
			return err
		}

		stack := stackWaiter.AwaitFinalStatus(envStackName)
		if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}
}

func (workflow *environmentWorkflow) environmentRolesetTerminator(rolesetDeleter common.RolesetDeleter, environmentName string) Executor {
	return func() error {
		err := rolesetDeleter.DeleteEnvironmentRoleset(environmentName)
		if err != nil {
			return err
		}
		return nil
	}
}

func (workflow *environmentWorkflow) environmentKubernetesIngressTerminator(environmentName string) Executor {
	return func() error {
		log.Noticef("Terminating ingress in environment '%s'", environmentName)

		err := workflow.kubernetesResourceManager.DeleteResource("v1", "Namespace", "", "mu-ingress")
		if err != nil {
			log.Warningf("Unable to delete namespace 'mu-ingress': %s", err)
		}
		return nil
	}
}

func (workflow *environmentWorkflow) environmentElbTerminator(namespace string, environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Terminating ELB environment '%s' ...", environmentName)
		envStackName := common.CreateStackName(namespace, common.StackTypeLoadBalancer, environmentName)
		err := stackDeleter.DeleteStack(envStackName)
		if err != nil {
			return err
		}

		stack := stackWaiter.AwaitFinalStatus(envStackName)
		if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}
}
func (workflow *environmentWorkflow) environmentVpcTerminator(namespace string, environmentName string, stackDeleter common.StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		log.Noticef("Terminating VPC environment '%s' ...", environmentName)
		vpcStackName := common.CreateStackName(namespace, common.StackTypeVpc, environmentName)
		err := stackDeleter.DeleteStack(vpcStackName)
		if err != nil {
			log.Debugf("Unable to delete VPC, but ignoring error: %v", err)
		}

		stack := stackWaiter.AwaitFinalStatus(vpcStackName)
		if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		targetStackName := common.CreateStackName(namespace, common.StackTypeTarget, environmentName)
		err = stackDeleter.DeleteStack(targetStackName)
		if err != nil {
			log.Debugf("Unable to delete VPC target, but ignoring error: %v", err)
		}

		stack = stackWaiter.AwaitFinalStatus(targetStackName)
		if stack != nil && !strings.HasSuffix(stack.Status, "_COMPLETE") {
			return fmt.Errorf("Ended in failed status %s %s", stack.Status, stack.StatusReason)
		}

		return nil
	}
}
