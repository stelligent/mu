[![Build Status](https://circleci.com/gh/stelligent/mu.svg?style=shield)](https://circleci.com/gh/stelligent/mu) [![Join the chat at https://gitter.im/stelligent/mu](https://badges.gitter.im/stelligent/mu.svg)](https://gitter.im/stelligent/mu?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/github.com/stelligent/mu)](https://goreportcard.com/report/github.com/stelligent/mu)


# Why?
Amazon ECS (EC2 Container Service) provides an excellent platform for deploying microservices as containers.  The challenge however is that there is a significant learning curve for microservice developers to deploy their applications in an efficient manner.  Specifically, they must learn to use CloudFormation to orchestrate the management of ECS, ECR, EC2, ELB, VPC, and IAM resources.  Additionally, tools like CodeBuild and CodePipeline must be mastered to create a continuous delivery pipeline for their microservices.

To address these challenges, this tool was created to simplify the declaration and administration of the AWS resources necessary to support microservices.  Similar to how the [Serverless Framework](https://serverless.com/) improved the developer experience of Lambda and API Gateway, this tool makes it easier for developers to use ECS as a microservices platform.

The `mu` tool uses CloudFormation stacks to manage all resources it creates.  Additionally, `mu` will not create any databases or other AWS resources to support itself.  It will only create resources (via CloudFormation) necessary to run your microservices.  This means at any point you can stop using `mu` and continue to manage the AWS resources that it created via AWS tools such as the CLI or the console.

![Architecture Diagram](https://github.com/stelligent/mu/wiki/img/ms-architecture-3.png)

# Get Started!
Install latest version to /usr/local/bin (or for additional options, see [installation options](https://github.com/stelligent/mu/wiki/Installation)):

```bash
curl -s https://raw.githubusercontent.com/stelligent/mu/master/install.sh | sh
```


# What's next?
Check out the [examples](https://github.com/stelligent/mu/wiki/Examples) to see common `mu.yml` configuration use cases:

* **[Basic](examples/basic)** - Simple website with continuous delivery pipeline deploying to dev and prod environments
* **[Test Automation](examples/pipeline-newman)** - Example of automating end-to-end testing via [Newman](https://github.com/postmanlabs/newman)
* **[Env Variables](examples/service-env-vars)** - Defining environment variables for the service
* **[HTTPS](examples/elb-https)** - Enable HTTPS on the ALB for an environment
* **[DNS](examples/elb-dns)** - Associate Route53 resource record with ALB for an environment
* **[VPC Target](examples/vpc-target)** - Targeting an existing VPC for an environment
* **[Custom CloudFormation](examples/custom-cloudformation)** - Demonstration of adding custom AWS resources via CloudFormation

Refer to the [wiki](https://github.com/stelligent/mu/wiki/Reference) for complete details on the configuration of `mu.yml` and the cli usage:

* **[Environments](https://github.com/stelligent/mu/wiki/Environments)** - managing VPCs, ECS clusters, container instances and ALBs
* **[Services](https://github.com/stelligent/mu/wiki/Services)** - managing ECS service configuration
* **[Pipelines](https://github.com/stelligent/mu/wiki/Pipelines)** - managing continuous delivery pipelines

# Contributing

Want to contribute to Mu?  Awesome!  Check out the [contributing guidelines](CONTRIBUTING.md) to get involved.

## Building from source

* Install Go tools 1.7+ - (https://golang.org/doc/install)
* Install [Glide](https://github.com/Masterminds/glide) via `curl https://glide.sh/get | sh`
* Clone this repo `git clone git@github.com:stelligent/mu.git $GOPATH/src/github.com/stelligent/mu`
* Go to src `cd $GOPATH/src/github.com/stelligent/mu`
* Build with `make`
