package workflows

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"io"
	"sync"
)

type purgeWorkflow struct {
	repoName string
}

// NewPurge create a new workflow for purging mu resources
func NewPurge(ctx *common.Context, writer io.Writer) Executor {

	workflow := new(purgeWorkflow)

	return newPipelineExecutor(
		workflow.purgeWorker(ctx, ctx.StackManager, writer),
	)
}

//
//
// main.go: main
// cli/app.go: NewApp
// cli/environments.go: newEnvironmentsCommand
// cli/environments.go: newEnvironmentsTerminateCommand
// workflows/environment_terminate.go: NewEnvironmentTerminate

// main.go: main
// cli/app.go: NewApp
// cli/services.go: newServicesCommand
// cli/services.go: newServicesUndeployCommand
// workflows/service_undeploy.go: newServiceUndeployer

//Workflow sequence
//
//for region in region-list (default to current, maybe implement a --region-list or --all-regions switch)
//  for namespace in namespaces (default to specified namespace)
//    for environment in all-environments (i.e. acceptance/production)
//      for service in services (all services in environment)
//         invoke 'svc undeploy'
//         invoke `env term`
//remove ECS repo
//invoke `pipeline term`
//remove s3 bucket containing environment name
//remove RDS databases
//
//other artifacts to remove:
//* common IAM roles
//* cloudwatch buckets
//* cloudwatch dashboards
//* (should be covered by CFN stack removal)
//* ECS scheduled tasks
//* SES
//* SNS
//* SQS
//* ELB
//* EC2 subnet
//* EC2 VPC Gateway attachment
//* security groups
//* EC2 Network ACL
//* EC2 Routetable
//* CF stacks

func removeStacksByStatus(stacks []*common.Stack, statuses []string) []*common.Stack {
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

func (workflow *purgeWorkflow) purgeWorker(ctx *common.Context, stackLister common.StackLister, writer io.Writer) Executor {
	return func() error {

		// TODO establish outer loop for regions
		// TODO establish outer loop for multiple namespaces
		purgeMap := make(map[common.StackType][]*common.Stack)

		// gather all the stackNames for each type (in parallel)
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(common.AllStackTypes))
		for _, stackType := range common.AllStackTypes {
			go func(sType common.StackType) {
				defer waitGroup.Done()
				stackNames, err := stackLister.ListStacks(sType)
				if err != nil {
					log.Warning("couldn't list stacks of type %V", sType)
				}
				stackNames = removeStacksByStatus(stackNames, []string{cloudformation.StackStatusRollbackComplete})
				purgeMap[sType] = stackNames
			}(stackType)
		}
		waitGroup.Wait()

		//environmentNames, err := stackLister.ListStacks(common.StackTypeEnv)
		//if err != nil {
		//	return err
		//}
		//environmentNames = removeStacksByStatus(environmentNames, []string{ cloudformation.StackStatusRollbackComplete})
		//purgeMap["env"] = environmentNames
		//
		//services, err := stackLister.ListStacks(common.StackTypeService)
		//if err != nil {
		//	return err
		//}
		//purgeMap["services"] = services
		//
		//databases, err := stackLister.ListStacks(common.StackTypeDatabase)
		//if err != nil {
		//	return err
		//}
		//purgeMap["databases"] = databases
		//
		//repos, err := stackLister.ListStacks(common.StackTypeRepo)
		//if err != nil {
		//	return err
		//}
		//purgeMap["repos"] = repos
		//
		//buckets, err := stackLister.ListStacks(common.StackTypeBucket)
		//if err != nil {
		//	return err
		//}
		//purgeMap["buckets"] = buckets
		//
		//consuls, err := stackLister.ListStacks(common.StackTypeConsul)
		//if err != nil {
		//	return err
		//}
		//purgeMap["consuls"] = consuls
		//
		//apps, err := stackLister.ListStacks(common.StackTypeApp)
		//if err != nil {
		//	return err
		//}
		//purgeMap["app"] = apps
		//
		//roles, err := stackLister.ListStacks(common.StackTypeIam)
		//if err != nil {
		//	return err
		//}
		//roles = removeStacksByStatus(roles, []string{ cloudformation.StackStatusRollbackComplete})
		//purgeMap["roles"] = roles
		//
		//elbs, err := stackLister.ListStacks(common.StackTypeLoadBalancer)
		//if err != nil {
		//	return err
		//}
		//purgeMap["elbs"] = elbs
		//
		//schedules, err := stackLister.ListStacks(common.StackTypeSchedule)
		//if err != nil {
		//	return err
		//}
		//purgeMap["schedules"] = schedules
		//
		//targets, err := stackLister.ListStacks(common.StackTypeTarget)
		//if err != nil {
		//	return err
		//}
		//purgeMap["targets"] = targets
		//
		//vpcs, err := stackLister.ListStacks(common.StackTypeVpc)
		//if err != nil {
		//	return err
		//}
		//purgeMap["vpcs"] = vpcs
		//
		//codePipelines, err := stackLister.ListStacks(common.StackTypePipeline)
		//if err != nil {
		//	return err
		//}
		//purgeMap["pipelines"] = codePipelines

		table := CreateTableSection(writer, PurgeHeader)
		stackCount := 0
		for stackType, stackList := range purgeMap {
			for _, stack := range stackList {
				log.Infof("stackType %v, stack %v", stackType, stack)
				table.Append([]string{
					Bold(stackType),
					stack.Name,
					fmt.Sprintf(KeyValueFormat, colorizeStackStatus(stack.Status), stack.StatusReason),
					stack.StatusReason,
					stack.LastUpdateTime.Local().Format(LastUpdateTime),
				})
				stackCount++
			}
		}
		table.Render()

		// create a grand master list of all the things we're going to delete
		var executors []Executor

		// add the services we're going to terminate
		svcWorkflow := new(serviceWorkflow)
		for _, service := range purgeMap[common.StackTypeService] {
			executors = append(executors, svcWorkflow.serviceInput(ctx, service.Name))
			executors = append(executors, svcWorkflow.serviceUndeployer(ctx.Config.Namespace, "", ctx.StackManager, ctx.StackManager))
		}

		// Add the terminator jobs to the master list for each environment
		envWorkflow := new(environmentWorkflow)
		for _, environmentName := range purgeMap[common.StackTypeEnv] {
			// Add the terminator jobs to the master list for each service
			executors = append(executors, envWorkflow.environmentServiceTerminator(environmentName.Name, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.RolesetManager))
			executors = append(executors, envWorkflow.environmentDbTerminator(environmentName.Name, ctx.StackManager, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentEcsTerminator(ctx.Config.Namespace, environmentName.Name, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentConsulTerminator(ctx.Config.Namespace, environmentName.Name, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentRolesetTerminator(ctx.RolesetManager, environmentName.Name))
			executors = append(executors, envWorkflow.environmentElbTerminator(ctx.Config.Namespace, environmentName.Name, ctx.StackManager, ctx.StackManager))
			executors = append(executors, envWorkflow.environmentVpcTerminator(ctx.Config.Namespace, environmentName.Name, ctx.StackManager, ctx.StackManager))
		}

		// add the pipelines to terminate (one for each service?)
		codePipelineWorkflow := new(pipelineWorkflow)
		for _, codePipeline := range purgeMap[common.StackTypePipeline] {
			executors = append(executors, codePipelineWorkflow.serviceFinder(codePipeline.Name, ctx))
			executors = append(executors, codePipelineWorkflow.pipelineTerminator(ctx.Config.Namespace, ctx.StackManager, ctx.StackManager))
			executors = append(executors, codePipelineWorkflow.pipelineRolesetTerminator(ctx.RolesetManager))
		}

		log.Infof("total of %d stacks of %d types to purge", stackCount, len(executors))

		// newPipelineExecutorNoStop is just like newPipelineExecutor, except that it doesn't stop on error
		executor := newPipelineExecutorNoStop(executors...)

		// run everything we've collected
		executor()
		return nil
	}
}
