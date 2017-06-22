* consider CodeDeploy blue/green
* ecs? (cleanup old services...new target group and service...then switch)
* provision ec2 ASG for `svc deploy` (cleanup old groups...new target group & asg...then codedeploy....then switch)
* pipeline - skip image (maybe create new revision??)
* dont show ecs for `env show` if type is ec2
* dont show ecs for `svc show` if type is ec2
