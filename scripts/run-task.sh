#!/usr/bin/env bash

if [ ${#} -lt 2 ]; then
    echo "Usage: $0 [mu-service-stack-name] [command]"
    exit 1
fi

MU_SERVICE_STACK=$1
CMD=
for i in "${@:2}"; do
    if [ ! -z ${CMD} ]; then
        CMD="$CMD,"
    fi
    CMD="$CMD\"$i\""
done

ECS_SERVICE_NAME=$(aws cloudformation describe-stacks --stack-name ${MU_SERVICE_STACK} --query "Stacks[0].Parameters[?ParameterKey=='ServiceName'].ParameterValue" --output text)
ECS_TASK_DEFINITION=$(aws cloudformation describe-stacks --stack-name ${MU_SERVICE_STACK} --query "Stacks[0].Outputs[?OutputKey=='MicroserviceTaskDefinition'].OutputValue" --output text)
ECS_CLUSTER=$(aws cloudformation describe-stacks --stack-name ${MU_SERVICE_STACK} --query "Stacks[0].Outputs[?OutputKey=='EcsCluster'].OutputValue" --output text)


OVERRIDES="
{
  \"containerOverrides\": [
    {
      \"name\": \"${ECS_SERVICE_NAME}\",
      \"command\": [${CMD}]
    }
  ]
}
"

ECS_TASK_ARN=$(aws ecs run-task --cluster ${ECS_CLUSTER} --task-definition ${ECS_TASK_DEFINITION} --query "tasks[0].taskArn" --output text --overrides "${OVERRIDES}")
aws ecs wait tasks-running --cluster ${ECS_CLUSTER} --tasks ${ECS_TASK_ARN}
aws ecs wait tasks-stopped --cluster ${ECS_CLUSTER} --tasks ${ECS_TASK_ARN}
EXIT_CODE=$(aws ecs describe-tasks --cluster ${ECS_CLUSTER} --tasks ${ECS_TASK_ARN} --query "tasks[0].containers[0].exitCode" --output text)
REASON=$(aws ecs describe-tasks --cluster ${ECS_CLUSTER} --tasks ${ECS_TASK_ARN} --query "tasks[0].containers[0].reason" --output text)

if [ "$EXIT_CODE" != "0" ]; then
    echo "Error: ${REASON} (exit=${EXIT_CODE})"
    exit 1
fi

