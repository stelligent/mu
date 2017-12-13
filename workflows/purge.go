package workflows

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"io"
	"strings"
)

type purgeWorkflow struct {
	RepoName string
}
type stackTerminateWorkflow struct {
	Stack *common.Stack
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

func filterStacksByType(stacks []*common.Stack, stackType common.StackType) []*common.Stack {
	var ret []*common.Stack
	for _, stack := range stacks {
		if stack.Tags["type"] == string(stackType) {
			ret = append(ret, stack)
		}
	}
	return ret
}

func (workflow *stackTerminateWorkflow) stackTerminator(ctx *common.Context, stackDeleter common.StackDeleter, stackLister common.StackLister, ecrRepoDeleter common.EcrRepoDeleter, s3stackDeleter common.S3StackDeleter, stackWaiter common.StackWaiter) Executor {
	return func() error {
		// get any dependent resources
		resources, err := stackLister.GetResourcesForStack(workflow.Stack)
		log.Info("resources %V", resources)
		if err != nil {
			return err
		}
		// do pre-delete API calls here (like deleting files from S3 bucket, before trying to delete bucket)
		for _, resource := range resources {
			if *resource.ResourceType == "AWS::S3::Bucket" {
				fqBucketName := resource.PhysicalResourceId
				log.Debugf("delete bucket: fullname=%s", *fqBucketName)
				// empty the bucket first
				s3stackDeleter.DeleteS3BucketObjects(*fqBucketName)
			} else if *resource.ResourceType == "AWS::ECR::Repository" {
				log.Infof("ECR::Repository %V", resource.PhysicalResourceId)
				ecrRepoDeleter.DeleteImagesFromEcrRepo(*resource.PhysicalResourceId)
			} else {
				log.Infof("don't know how to delete a type %s", *resource.ResourceType)
			}
		}
		// delete the stack object
		err = stackDeleter.DeleteStack(workflow.Stack.Name)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				log.Errorf("%v", aerr.Error())
			} else {
				log.Errorf("%v", err)
			}
		}
		// wait for the result
		svcStack := stackWaiter.AwaitFinalStatus(workflow.Stack.Name)
		if svcStack != nil && !strings.HasSuffix(svcStack.Status, "_COMPLETE") {
			log.Errorf("Ended in failed status %s %s", svcStack.Status, svcStack.StatusReason)
		}

		// do post-delete API calls here (just in case anything was left over from the DeleteStack, abaove
		for _, resource := range resources {
			if *resource.ResourceType == "AWS::S3::Bucket" {
				fqBucketName := resource.PhysicalResourceId
				err2 := s3stackDeleter.DeleteS3Bucket(*fqBucketName)
				if err2 != nil {
					if aerr, ok := err2.(awserr.Error); ok {
						log.Errorf("couldn't delete S3 Bucket %s %v", *fqBucketName, aerr.Error())
					} else {
						log.Errorf("couldn't delete S3 Bucket %s %v", *fqBucketName, err2)
					}
				}
			}
		}
		return nil
	}
}

func (workflow *purgeWorkflow) purgeWorker(ctx *common.Context, stackLister common.StackLister, writer io.Writer) Executor {
	return func() error {

		// TODO establish outer loop for regions
		// TODO establish outer loop for multiple namespaces
		// purgeMap := make(map[string][]*common.Stack)

		// gather all the stackNames for each type (in parallel)
		stacks, err := stackLister.ListStacks(common.StackTypeAll)
		if err != nil {
			log.Warning("couldn't list stacks (all)")
		}
		stacks = removeStacksByStatus(stacks, []string{cloudformation.StackStatusRollbackComplete})

		table := CreateTableSection(writer, PurgeHeader)
		stackCount := 0
		for _, stack := range stacks {
			stackType, ok := stack.Tags["type"]
			if ok {
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

		// TODO - scheduled tasks are attached to service, so must be deleted first.
		// common.StackTypeSchedule

		svcWorkflow := new(serviceWorkflow)

		// add the services we're going to terminate

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

		// add the pipelines to terminate
		codePipelineWorkflow := new(pipelineWorkflow)
		for _, codePipeline := range filterStacksByType(stacks, common.StackTypePipeline) {
			// log.Infof("%s %v", codePipeline.Name, codePipeline.Tags)
			executors = append(executors, codePipelineWorkflow.serviceFinder(codePipeline.Tags["service"], ctx))
			executors = append(executors, codePipelineWorkflow.pipelineTerminator(ctx.Config.Namespace, ctx.StackManager, ctx.StackManager))
			executors = append(executors, codePipelineWorkflow.pipelineRolesetTerminator(ctx.RolesetManager))
		}

		// add the buckets to remove
		for _, bucket := range filterStacksByType(stacks, common.StackTypeBucket) {
			log.Infof("%s %v", bucket.Name, bucket.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = bucket
			//                                           (ctx *common.Context, stackDeleter common.StackDeleter, stackLister common.StackLister, ecrRepoDeleter common.EcrRepoDeleter, s3stackDeleter common.S3StackDeleter, stackWaiter common.StackWaiter) Executor {
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.StackManager))

			// func (workflow *stackTerminateWorkflow) stackTerminator(ctx *common.Context,  stackDeleter common.StackDeleter, stackLister common.StackLister, stackWaiter common.StackWaiter) Executor {
		}

		// add the buckets to remove
		for _, repo := range filterStacksByType(stacks, common.StackTypeRepo) {
			log.Infof("%s %v", repo.Name, repo.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = repo
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		// add the iam roles to delete
		for _, roleStack := range filterStacksByType(stacks, common.StackTypeIam) {
			log.Infof("%s %v", roleStack.Name, roleStack.Tags)
			workflow := new(stackTerminateWorkflow)
			workflow.Stack = roleStack
			executors = append(executors, workflow.stackTerminator(ctx, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.StackManager, ctx.StackManager))
		}

		// add the ecs repos to terminate

		// aws ecr describe-repositories
		//	"repositories": [
		//		"repositoryArn": "arn:aws:ecr:eu-west-1:324320755747:repository/mu-tim-mu-banana",
		//		"registryId": "324320755747",
		//		"repositoryName": "mu-tim-mu-banana",
		//		"repositoryUri": "324320755747.dkr.ecr.eu-west-1.amazonaws.com/mu-tim-mu-banana",
		//		"createdAt": 1511561499.0
		// aws ecr describe-images --repository-name mu-tim-mu-banana
		// aws ecr batch-delete-image --repository-name ubuntu --image-ids imageTag=precise

		// QUESTION: do we want to delete stacks of type CodeCommit?  (currently, my example is github)

		// common.StackTypeLoadBalancer
		// common.StackTypeDatabase - databaseWorkflow
		// common.StackTypeBucket
		// common.StackTypeVpc

		// logsWorkflow (for cloudwatch workflows)

		log.Infof("total of %d stacks of %d types to purge", stackCount, len(executors))

		// newPipelineExecutorNoStop is just like newPipelineExecutor, except that it doesn't stop on error
		executor := newPipelineExecutorNoStop(executors...)

		// run everything we've collected
		executor()
		return nil
	}
}
