#!/usr/bin/env python
import os
import subprocess
import sys
import boto.utils
import boto3
import requests
import datetime
import time


def get_contents(filename):
    with open(filename) as f:
        return f.read()


def get_ecs_introspection_url(resource):
    # 172.17.0.1 is the docker network bridge ip
    return 'http://172.17.0.1:51678/v1/' + resource


def contains_key(d, key):
    return key in d and d[key] is not None


def get_local_container_info():
    # get the docker container id
    # http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-agent-introspection.html
    docker_id = os.path.basename(get_contents("/proc/1/cpuset")).strip()

    if docker_id is None:
        raise Exception("Unable to find docker id")

    ecs_local_task = requests.get(get_ecs_introspection_url('tasks') + '?dockerid=' + docker_id).json()

    task_arn = ecs_local_task['Arn']

    if task_arn is None:
        raise Exception("Unable to find task arn for container %s in ecs introspection api" % docker_id)

    ecs_local_container = None

    if contains_key(ecs_local_task, 'Containers'):
        for c in ecs_local_task['Containers']:
            if c['DockerId'] == docker_id:
                ecs_local_container = c

    if ecs_local_container is None:
        raise Exception("Unable to find container %s in ecs introspection api" % docker_id)

    return ecs_local_container['Name'], task_arn

def get_container_ports(env_map,region):
    try:
        ecs_metadata = requests.get(get_ecs_introspection_url('metadata')).json()
        cluster = ecs_metadata['Cluster']
    except:
        return

    container_name, task_arn = get_local_container_info()

    # Get the container info from ECS. This will give us the port mappings
    ecs = boto3.client('ecs', region_name=region)
    response = ecs.describe_tasks(
        cluster=cluster,
        tasks=[
            task_arn,
        ]
    )

    task = None
    if contains_key(response, 'tasks'):
        for t in response['tasks']:
            if t['taskArn'] == task_arn:
                task = t

    if task is None:
        raise Exception("Unable to locate task %s" % task_arn)

    container = None
    if contains_key(task, 'containers'):
        for c in task['containers']:
            if c['name'] == container_name:
                container = c

    if container is None:
        raise Exception("Unable to find ecs container %s" % container_name)

    if contains_key(container, 'networkBindings'):
        for b in container['networkBindings']:
            key = ("HOST_PORT_%s_%d" % (b['protocol'].upper(), b['containerPort']))
            env_map[key] = ("%d" % (b['hostPort']))


def main():

    metadata = boto.utils.get_instance_metadata()
    region = metadata['placement']['availability-zone'][:-1]  # last char is the zone, which we don't care about

    env_map = dict(os.environ)
    env_map["HOST_IP"] = metadata['local-ipv4']
    env_map["HOST_NAME"] = metadata['local-hostname']
    get_container_ports(env_map, region)

    print(sys.argv[1:])
    os.execve('/bin/sh',['sh','-c', ' '.join(sys.argv[1:])],env_map)


if __name__ == '__main__':
    main()
