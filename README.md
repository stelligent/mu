[![Build Status](https://travis-ci.org/stelligent/mu.svg?branch=develop)](https://travis-ci.org/stelligent/mu)

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

## Commands

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
> mu service deploy <environment_name> [-s <service_name>]

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



