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



		// QUESTION: do we want to delete stacks of type CodeCommit?  (currently, my example is github)

		// common.StackTypeLoadBalancer
		// common.StackTypeDatabase - databaseWorkflow
		// common.StackTypeBucket
		// common.StackTypeVpc

		// logsWorkflow (for cloudwatch workflows)



		// TODO establish outer loop for regions
		// TODO establish outer loop for multiple namespaces

