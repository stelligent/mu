# Problem Statement
Amazon ECS (EC2 Container Service) provides an excellent platform for deploying microservices as containers.  The challenge however is that there is a significant learning curve for microservice developers to deploy their applications in an efficient manner.  Specifically, they must learn to use CloudFormation to orchestrate the management of ECS, ECR, EC2, ELB, VPC, and IAM resources.  Additionally, tools like CodeBuild and CodePipeline must be mastered to create a continuous delivery pipeline for their microservices.

To address these challenges, we will create a tool to simplify the declaration and administration of the AWS resources necessary to support microservices.  Similar to how the [Serverless Framework](https://serverless.com/) improved the developer experience of Lambda and API Gateway, this tool will make it easier for developers to use ECS as a microservices platform.

For more details on the intended architecture, see [Microservices Platform with ECS](https://stelligent.com/2016/10/06/microservices-platform-with-ecs/).

#Assumptions
1. **Polyglot** - There will be no prescribed language or framework for developing the microservices.  The only requirement will be that the service will be run inside a container and exposed via an HTTP endpoint.
2. **Cloud Provider** - At this point, the tool will assume AWS for the cloud provider and will not be written in a cloud agnostic manner.  However, this does not preclude refactoring to add support for other providers at a later time.
3. **Declarative** - All resource administration will be handled in a declarative vs. imperative manner.  A file will be used to declared the desired state of the resources and the tool will simply assert the actual state matches the desired state.  The tool will accomplish this by generating CloudFormation templates.
4. **Stateless** - The tool will not maintain its own state.  Rather, it will rely on the CloudFormation stacks to determine the state of the platform.
5. **Secure** - All security will be managed by AWS IAM credentials.  No additional authentication or authorization mechanisms will be introduced.
6. **Language** - TBD.  Need to determine the language to use for developing the tool.  Options in order of preference include Go, Node.js, Python.


#Capabilities
## Resource Declaration
A YAML file will be used to declare microservice resources.  There are two types of resources defined in the YAML file, environments and services.

Environments contain an ECS cluster, ECS container instances (with ASG), and an ALB.  Additionally, environments contain (or reference) a VPC.  A sample environment resource may look like:


```
-
environments:
  dev:
    loadbalancer:
      hostname: api-dev.example.com
    cluster:
      desiredCapacity: 1
      maxSize: 1
  production:
    loadbalancer:
      hostname: api.example.com
    cluster:
      desiredCapacity: 2
      maxSize: 5
```


Services contain an ECS service, ECS task, ALB target group and ECR.  Additionally service can contain an optional CodeBuild and CodePipeline resource for a CD pipeline.
```
-
service:
  desiredCount: 2
  pipeline:
    devEnvironment: dev
    prodEnvironment: production
```

## CLI
The majority of code for this tool will be to provide a CLI to manage CloudFormation stacks based on the resources declared in the YAML file.  The list of available commands are:

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


## Plugin
A plugin framework should be available for developer to contribute extensions for specific languages.  For example, a Java developer using Spring Boot should be able to use a Spring Boot plugin to define the Eureka, ConfigServer and Zuul router for their environment as follows:

```
environments:
  dev:
    loadbalancer:
      hostname: api-dev.example.com
    springboot:
      eureka:
        desiredCapacity: 1
      configServer:
        sourceUrl: https://github.com/example/configrepo
```

      
## UI
A web based user interface will be created to provide visibility into the resources in the platform.   The UI will allow a view into the list of pipelines, services, and environments defined in a given AWS account.  It will only provide read only access to the resources and will not provide ability to change the resources.

The UI will consist of an Angular2 application hosted in S3 with APIs in Lambda/API Gateway.  The UI will be secured via AWS credentials and Cognito. 


