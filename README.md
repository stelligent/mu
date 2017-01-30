[![Build Status](https://circleci.com/gh/stelligent/mu.svg?style=shield)](https://circleci.com/gh/stelligent/mu) [![Join the chat at https://gitter.im/stelligent/mu](https://badges.gitter.im/stelligent/mu.svg)](https://gitter.im/stelligent/mu?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)



# Why?
Amazon ECS (EC2 Container Service) provides an excellent platform for deploying microservices as containers.  The challenge however is that there is a significant learning curve for microservice developers to deploy their applications in an efficient manner.  Specifically, they must learn to use CloudFormation to orchestrate the management of ECS, ECR, EC2, ELB, VPC, and IAM resources.  Additionally, tools like CodeBuild and CodePipeline must be mastered to create a continuous delivery pipeline for their microservices.

To address these challenges, this tool was created to simplify the declaration and administration of the AWS resources necessary to support microservices.  Similar to how the [Serverless Framework](https://serverless.com/) improved the developer experience of Lambda and API Gateway, this tool makes it easier for developers to use ECS as a microservices platform.

For more details on the intended architecture, see [Microservices Platform with ECS](https://stelligent.com/2016/10/06/microservices-platform-with-ecs/).

# Installation

```bash
# Install latest version to /usr/local/bin
curl -s https://raw.githubusercontent.com/stelligent/mu/master/install.sh | sh

# Install v0.1.0 version to ~/bin
curl -s https://raw.githubusercontent.com/stelligent/mu/master/install.sh | INSTALL_VERSION=0.1.0 INSTALL_DIR=~/bin sh
```

# Commands

```
# List all environments
> mu env list

# Show details about a specific environment (ECS container instances, Running services, etc)
> mu env show <environment_name>

# Upsert an environment
> mu env up <environment_name>

# Terminate an environment
> mu env terminate <environment_name>

# Show details about a specific service (Which versions in which environments, pipeline status)
> mu service show [-s <service_name>]

# Deploy the service to an environment
> mu service deploy <environment_name>

# Set an environment variable(s) for a service
> mu service setenv <environment_name> [-s <service_name>] key=value[,...]

# Undeploy the service from an environment
> mu service undeploy <environment_name> [-s <service_name>]

# List the pipelines
> mu pipeline list

# Show the pipeline details for a specific service
> mu pipeline show <service_name>

# Upsert the pipeline
> mu pipeline up [-s <service_name>] [-u <repo_url>] [-t <repo_token>]

# Terminate the pipeline
> mu pipeline terminate [-s <service_name>]
```

# Configuration
The definition of your environments, services and pipelines is done via a YAML file (default `./mu.yml`).

```
---
### Region to utilize
region: us-west-2

### Define a list of environments
environments:

  # The unique name of the environment  (required)
  - name: dev

    ### Attributes for the ECS container instances
    cluster:
      imageId: ami-xxxxxx           # The AMI to use for the ECS container instances (default: latest ECS optimized AMI)
      instanceTenancy: default      # Whether to use default or dedicated tenancy (default: default)
      desiredCapacity: 1            # Desired number of ECS container instances (default 1)
      maxSize: 2                    # Max size to scale the ECS ASG to (default: 2)
      keyName: my-keypair           # name of EC2 keypair to associate with ECS container instances (default: none)
      sshAllow: 0.0.0.0/0           # CIDR block to allow SSH access from (default: 0.0.0.0/0)
      scaleOutThreshold: 80         # Threshold for % memory utilization to scale out ECS container instances (default: 80)
      scaleInThreshold: 30          # Threshold for % memory utilization to scale in ECS container instances (default: 30)

    ### attributes for the VPC to target.  If not defined, a VPC will be created. (default: none)
    vpcTarget:
        vpcId: vpc-xxxxx            # The id of the VPC to launch ECS container instances into
        publicSubnetIds:            # The list of subnets to use for ECS container instances
          - subnet-xxxxx
          - subnet-xxxxy
          - subnet-xxxxz

### Define the service for this repo
service:
  name: my-service                   # The unique name of the service (default: the name of the directory that mu.yml was in)
  desiredCount: 4                    # The desired number of tasks to run for the service (default: 2)
  dockerfile: ./Dockerfile           # The relative path to the Dockerfile to build images (default: ./Dockerfile)
  imageRepository: tutum/hello-world # The repository to push images to and deploy services from.  Leave unset to have mu manage an ECR repository (default: none)
  port: 80                           # The port to expose from the container (default: 8080)
  healthEndpoint: /health            # The endpoint inside the container to determine if the task is healthy (default: /health)
  cpu: 20                            # The number of CPU units to allocate to each task (default: 10)
  memory: 400                        # The amount of memory in MiB to allocate to each task (default: 300)

  # The paths to match on in the ALB and route to this service.  Leave blank to not create an ALB target group for this service (default: none)
  pathPatterns:
    - /bananas
    - /apples
```


# Contributing

Want to contribute to Mu?  Awesome!  Check out the [contributing guidelines](CONTRIBUTING.md) to get involved.

## Building from source

* Install Go tools 1.7+ - (https://golang.org/doc/install)
* Install [Glide](https://github.com/Masterminds/glide) via `curl https://glide.sh/get | sh`
* Clone this repo `git clone git@github.com:stelligent/mu.git $GOPATH/src/github.com/stelligent/mu`
* Go to src `cd $GOPATH/src/github.com/stelligent/mu`
* Build with `make`
