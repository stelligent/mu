# Change Log

## [v1.2.1](https://github.com/stelligent/mu/tree/v1.2.1) (2017-12-07)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.1.3...v1.2.1)

**Implemented enhancements:**

- Add environment provider for ECS+Fargate [\#234](https://github.com/stelligent/mu/issues/234)

**Fixed bugs:**

- Mu files with different names are ignored when generating env.json [\#236](https://github.com/stelligent/mu/issues/236)

**Merged pull requests:**

- Release 1.2.1 [\#239](https://github.com/stelligent/mu/pull/239) ([cplee](https://github.com/cplee))
- add support for fargate as an environment provider [\#238](https://github.com/stelligent/mu/pull/238) ([cplee](https://github.com/cplee))

## [v1.1.3](https://github.com/stelligent/mu/tree/v1.1.3) (2017-12-05)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.1.2...v1.1.3)

**Implemented enhancements:**

- Externalize S3 Bucket creation [\#216](https://github.com/stelligent/mu/issues/216)
- Dynamic variables in mu.yml [\#209](https://github.com/stelligent/mu/issues/209)
- Add option to set the http proxy the SDK will use [\#223](https://github.com/stelligent/mu/pull/223) ([mince27](https://github.com/mince27))

**Merged pull requests:**

- Fix for issue 236: Added -c flag to env show command to generate env.… [\#237](https://github.com/stelligent/mu/pull/237) ([akuma12](https://github.com/akuma12))
- Feature/externalize s3 bucket [\#226](https://github.com/stelligent/mu/pull/226) ([cplee](https://github.com/cplee))
- Issue 209 documentation \(environment variable substitution\) [\#225](https://github.com/stelligent/mu/pull/225) ([timbaileyjones](https://github.com/timbaileyjones))
- Issue 209 dynamic variables [\#224](https://github.com/stelligent/mu/pull/224) ([timbaileyjones](https://github.com/timbaileyjones))

## [v1.1.2](https://github.com/stelligent/mu/tree/v1.1.2) (2017-11-16)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.1.1...v1.1.2)

**Fixed bugs:**

- Issue with mu env -f json [\#221](https://github.com/stelligent/mu/issues/221)

**Merged pull requests:**

- fix \#221 [\#222](https://github.com/stelligent/mu/pull/222) ([cplee](https://github.com/cplee))

## [v1.1.1](https://github.com/stelligent/mu/tree/v1.1.1) (2017-11-15)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.0.6...v1.1.1)

**Implemented enhancements:**

- Allow custom AMIs based on OS other than amazon linux to work [\#205](https://github.com/stelligent/mu/issues/205)
- Update environment ASG to use target tracking policies [\#177](https://github.com/stelligent/mu/issues/177)
- Add S3 Source support for CodePipeline [\#176](https://github.com/stelligent/mu/issues/176)
- Service autoscaling [\#171](https://github.com/stelligent/mu/issues/171)
- Allow referencing external files for templates [\#167](https://github.com/stelligent/mu/issues/167)
- Nested mu.yml files [\#162](https://github.com/stelligent/mu/issues/162)
- Scheduled tasks [\#158](https://github.com/stelligent/mu/issues/158)
- Issue 158 scheduled tasks [\#217](https://github.com/stelligent/mu/pull/217) ([timbaileyjones](https://github.com/timbaileyjones))
- Template splice [\#215](https://github.com/stelligent/mu/pull/215) ([cplee](https://github.com/cplee))
- issue-176 Added S3 Source option to CodePipeline [\#198](https://github.com/stelligent/mu/pull/198) ([akuma12](https://github.com/akuma12))

**Fixed bugs:**

- Unable to specify branch [\#201](https://github.com/stelligent/mu/issues/201)
- Bug with latest Consul release [\#194](https://github.com/stelligent/mu/issues/194)
- Infinite loop trying to find .git on Windows [\#189](https://github.com/stelligent/mu/issues/189)

**Merged pull requests:**

- v1.1.1 [\#218](https://github.com/stelligent/mu/pull/218) ([cplee](https://github.com/cplee))
- Issue 171 and 177 [\#213](https://github.com/stelligent/mu/pull/213) ([cplee](https://github.com/cplee))
- Extensions [\#210](https://github.com/stelligent/mu/pull/210) ([cplee](https://github.com/cplee))
- Only install cfn-nag if needed. Add GOPATH to PATH [\#207](https://github.com/stelligent/mu/pull/207) ([jeremyhahn](https://github.com/jeremyhahn))
- add conditionals to the userdata and cfn-init based on os type [\#206](https://github.com/stelligent/mu/pull/206) ([cplee](https://github.com/cplee))
- Issue 201 [\#202](https://github.com/stelligent/mu/pull/202) ([cplee](https://github.com/cplee))
- Fixed git.go to walk up Windows dirs [\#199](https://github.com/stelligent/mu/pull/199) ([ataylor05](https://github.com/ataylor05))
- Asg 177 [\#188](https://github.com/stelligent/mu/pull/188) ([danielc2013](https://github.com/danielc2013))

## [v1.0.6](https://github.com/stelligent/mu/tree/v1.0.6) (2017-10-17)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.0.5...v1.0.6)

## [v1.0.5](https://github.com/stelligent/mu/tree/v1.0.5) (2017-10-16)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.0.4...v1.0.5)

## [v1.0.4](https://github.com/stelligent/mu/tree/v1.0.4) (2017-10-13)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.0.3...v1.0.4)

**Fixed bugs:**

- Quotes in RoleNames and !Sub in -iam.yml templates break when custom CloudFormation added to those stacks [\#192](https://github.com/stelligent/mu/issues/192)
- mu-iam-common stack fails if it has already been deployed in another region [\#187](https://github.com/stelligent/mu/issues/187)

**Merged pull requests:**

- issue-192 Removed single quotes around RoleNames and changed !Sub ent… [\#193](https://github.com/stelligent/mu/pull/193) ([akuma12](https://github.com/akuma12))

## [v1.0.3](https://github.com/stelligent/mu/tree/v1.0.3) (2017-10-11)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.0.2...v1.0.3)

**Implemented enhancements:**

- Distribute mu via homebrew [\#132](https://github.com/stelligent/mu/issues/132)

**Fixed bugs:**

- pipeline-iam.yml relies on ${namespace}-bucket-codedeploy Export which may not exist [\#186](https://github.com/stelligent/mu/issues/186)

**Merged pull requests:**

- fix for \#187 [\#191](https://github.com/stelligent/mu/pull/191) ([cplee](https://github.com/cplee))
- Issue \#187 - Add namespace to buckets and regions to IAM roles to avo… [\#190](https://github.com/stelligent/mu/pull/190) ([cplee](https://github.com/cplee))
- Added a formula Makefile target.  Resolves \#132. [\#183](https://github.com/stelligent/mu/pull/183) ([juddmon](https://github.com/juddmon))

## [v1.0.2](https://github.com/stelligent/mu/tree/v1.0.2) (2017-10-09)
[Full Changelog](https://github.com/stelligent/mu/compare/v1.0.1...v1.0.2)

**Fixed bugs:**

- Custom Cloudformation breaks the Pipeline template [\#181](https://github.com/stelligent/mu/issues/181)

**Closed issues:**

- Unable to create pipeline with 1.X [\#184](https://github.com/stelligent/mu/issues/184)

**Merged pull requests:**

- Bug fixes [\#185](https://github.com/stelligent/mu/pull/185) ([cplee](https://github.com/cplee))
- Resolve issue with invalid CFN for custom pipeline templates [\#182](https://github.com/stelligent/mu/pull/182) ([cplee](https://github.com/cplee))

## [v1.0.1](https://github.com/stelligent/mu/tree/v1.0.1) (2017-10-02)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.2.5...v1.0.1)

**Closed issues:**

- Rolling restart [\#164](https://github.com/stelligent/mu/issues/164)
- cfn\_nag [\#163](https://github.com/stelligent/mu/issues/163)
- Add ability to set/get the DB password [\#140](https://github.com/stelligent/mu/issues/140)
- Custom tags on stacks [\#88](https://github.com/stelligent/mu/issues/88)
- Customize stack prefix for namespacing [\#87](https://github.com/stelligent/mu/issues/87)
- Pass existing IAM roles and/or Security Groups [\#86](https://github.com/stelligent/mu/issues/86)
- Target multiple regions/accounts for environments [\#84](https://github.com/stelligent/mu/issues/84)
- Cleanup IAM policies [\#82](https://github.com/stelligent/mu/issues/82)

**Merged pull requests:**

- v1.0.1 [\#180](https://github.com/stelligent/mu/pull/180) ([cplee](https://github.com/cplee))
- IAM refactor [\#179](https://github.com/stelligent/mu/pull/179) ([cplee](https://github.com/cplee))
- Custom tags 88 [\#175](https://github.com/stelligent/mu/pull/175) ([danielc2013](https://github.com/danielc2013))
- Add namespace for stacks [\#174](https://github.com/stelligent/mu/pull/174) ([juddmon](https://github.com/juddmon))
- Rolling restart 164 [\#173](https://github.com/stelligent/mu/pull/173) ([danielc2013](https://github.com/danielc2013))
- Get and set db password [\#168](https://github.com/stelligent/mu/pull/168) ([juddmon](https://github.com/juddmon))

## [v0.2.5](https://github.com/stelligent/mu/tree/v0.2.5) (2017-08-16)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.2.4...v0.2.5)

**Merged pull requests:**

- release 0.2.5 [\#161](https://github.com/stelligent/mu/pull/161) ([cplee](https://github.com/cplee))

## [v0.2.4](https://github.com/stelligent/mu/tree/v0.2.4) (2017-08-10)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.2.3...v0.2.4)

**Merged pull requests:**

- release 0.2.4 [\#160](https://github.com/stelligent/mu/pull/160) ([cplee](https://github.com/cplee))

## [v0.2.3](https://github.com/stelligent/mu/tree/v0.2.3) (2017-07-26)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.2.2...v0.2.3)

**Closed issues:**

- Skip build/image step in pipeline [\#151](https://github.com/stelligent/mu/issues/151)
- Create Route53 resource record per service [\#150](https://github.com/stelligent/mu/issues/150)
- Support ALB routing via hostname [\#136](https://github.com/stelligent/mu/issues/136)

**Merged pull requests:**

- v0.2.3 [\#155](https://github.com/stelligent/mu/pull/155) ([cplee](https://github.com/cplee))
- issues for v0.2.3 [\#154](https://github.com/stelligent/mu/pull/154) ([cplee](https://github.com/cplee))

## [v0.2.2](https://github.com/stelligent/mu/tree/v0.2.2) (2017-07-24)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.2.1...v0.2.2)

**Merged pull requests:**

- v0.2.2 [\#153](https://github.com/stelligent/mu/pull/153) ([cplee](https://github.com/cplee))

## [v0.2.1](https://github.com/stelligent/mu/tree/v0.2.1) (2017-07-20)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.14...v0.2.1)

**Fixed bugs:**

- Unable to determine git revision  [\#142](https://github.com/stelligent/mu/issues/142)

**Closed issues:**

- Support alternate mu.yml files [\#146](https://github.com/stelligent/mu/issues/146)
- Support defining ec2 as the provider for an environment [\#144](https://github.com/stelligent/mu/issues/144)

**Merged pull requests:**

- 0.2.1 [\#152](https://github.com/stelligent/mu/pull/152) ([cplee](https://github.com/cplee))
- Issue 144 [\#149](https://github.com/stelligent/mu/pull/149) ([cplee](https://github.com/cplee))
- fix \#142 remove gogit dependency in place of manually opening file [\#148](https://github.com/stelligent/mu/pull/148) ([cplee](https://github.com/cplee))

## [v0.1.14](https://github.com/stelligent/mu/tree/v0.1.14) (2017-06-22)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.13...v0.1.14)

**Fixed bugs:**

- CodeCommit defects [\#134](https://github.com/stelligent/mu/issues/134)
- Rollback on pipeline command [\#105](https://github.com/stelligent/mu/issues/105)
- Auto increment ALB ListenerRule Priority [\#85](https://github.com/stelligent/mu/issues/85)
- Service build has redundant logs from docker [\#81](https://github.com/stelligent/mu/issues/81)

**Closed issues:**

- Variable substitution with single quotes in templates results in failure  [\#138](https://github.com/stelligent/mu/issues/138)
- How to configure desired capacity for consul server? [\#137](https://github.com/stelligent/mu/issues/137)
- View ECS cluster instances for a given service [\#131](https://github.com/stelligent/mu/issues/131)
- Progress output for CodeBuild [\#120](https://github.com/stelligent/mu/issues/120)
- Implement Service Discovery [\#116](https://github.com/stelligent/mu/issues/116)
- Determine service name from GitHub repo name [\#112](https://github.com/stelligent/mu/issues/112)
- Add support for CodeCommit [\#107](https://github.com/stelligent/mu/issues/107)
- Improve process for running commands after deployment [\#104](https://github.com/stelligent/mu/issues/104)
- Pipeline manages environment [\#102](https://github.com/stelligent/mu/issues/102)
- Tag CloudFormation stacks with commit id of mu.yml [\#99](https://github.com/stelligent/mu/issues/99)
- Extend/Override Generated CloudFormation  [\#97](https://github.com/stelligent/mu/issues/97)
- Version numbers for docker image [\#80](https://github.com/stelligent/mu/issues/80)
- View build logs [\#79](https://github.com/stelligent/mu/issues/79)
- Execute tests from CodePipeline [\#78](https://github.com/stelligent/mu/issues/78)
- Service Log Aggregation [\#77](https://github.com/stelligent/mu/issues/77)
- VPC enhancements [\#76](https://github.com/stelligent/mu/issues/76)
- Cleanup VPC reference [\#63](https://github.com/stelligent/mu/issues/63)
- Add support for HTTPS and DNS [\#62](https://github.com/stelligent/mu/issues/62)
- Vendoring of dependencies [\#60](https://github.com/stelligent/mu/issues/60)
- Blog post announcing mu [\#46](https://github.com/stelligent/mu/issues/46)
- Documentation for installing/using mu [\#43](https://github.com/stelligent/mu/issues/43)
- Terminate pipeline [\#31](https://github.com/stelligent/mu/issues/31)
- Add pipeline status to Show service  [\#30](https://github.com/stelligent/mu/issues/30)
- Show pipeline details [\#29](https://github.com/stelligent/mu/issues/29)
- List pipelines [\#28](https://github.com/stelligent/mu/issues/28)
- Create pipeline \(CodePipeline and CodeBuild\) [\#27](https://github.com/stelligent/mu/issues/27)
- Undeploy service [\#26](https://github.com/stelligent/mu/issues/26)
- Set service environment variable [\#25](https://github.com/stelligent/mu/issues/25)
- Add service status to show environment [\#24](https://github.com/stelligent/mu/issues/24)
- Show services [\#23](https://github.com/stelligent/mu/issues/23)
- Create service \(ECS task, ECS service, ALB target group\) [\#22](https://github.com/stelligent/mu/issues/22)
- Terminate environment [\#21](https://github.com/stelligent/mu/issues/21)
- Show environment [\#20](https://github.com/stelligent/mu/issues/20)
- List environments [\#19](https://github.com/stelligent/mu/issues/19)
- Create environment \(VPC, ECS cluster, ASG/LC for ECS instances, ALB\) [\#18](https://github.com/stelligent/mu/issues/18)
- Parse mu.yml file into domain objects [\#17](https://github.com/stelligent/mu/issues/17)

**Merged pull requests:**

- 0.1.4 - bug fixes [\#143](https://github.com/stelligent/mu/pull/143) ([cplee](https://github.com/cplee))
- fixes \#138 [\#139](https://github.com/stelligent/mu/pull/139) ([sumitsarkar](https://github.com/sumitsarkar))

## [v0.1.13](https://github.com/stelligent/mu/tree/v0.1.13) (2017-05-12)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.12...v0.1.13)

**Merged pull requests:**

- Release 0.1.13 [\#135](https://github.com/stelligent/mu/pull/135) ([cplee](https://github.com/cplee))
- Added a table to display the container details for service and env sh… [\#133](https://github.com/stelligent/mu/pull/133) ([jblouse](https://github.com/jblouse))
- Issue 104 [\#130](https://github.com/stelligent/mu/pull/130) ([jblouse](https://github.com/jblouse))

## [v0.1.12](https://github.com/stelligent/mu/tree/v0.1.12) (2017-05-03)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.11...v0.1.12)

**Closed issues:**

- Use param store for credentials [\#126](https://github.com/stelligent/mu/issues/126)
- Define databases in mu.yml [\#101](https://github.com/stelligent/mu/issues/101)

**Merged pull requests:**

- Release 0.1.12 [\#129](https://github.com/stelligent/mu/pull/129) ([cplee](https://github.com/cplee))
- Add support for databases to mu [\#127](https://github.com/stelligent/mu/pull/127) ([cplee](https://github.com/cplee))

## [v0.1.11](https://github.com/stelligent/mu/tree/v0.1.11) (2017-04-19)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.10...v0.1.11)

**Merged pull requests:**

- Use STDERR for warnings [\#125](https://github.com/stelligent/mu/pull/125) ([cplee](https://github.com/cplee))

## [v0.1.10](https://github.com/stelligent/mu/tree/v0.1.10) (2017-04-10)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.9...v0.1.10)

**Merged pull requests:**

- v0.10.0 final [\#124](https://github.com/stelligent/mu/pull/124) ([cplee](https://github.com/cplee))
- Issue 120 [\#122](https://github.com/stelligent/mu/pull/122) ([cplee](https://github.com/cplee))
- disable progess spinner if not running in terminal [\#121](https://github.com/stelligent/mu/pull/121) ([cplee](https://github.com/cplee))
- Add tags to CloudFormation stacks for repo name and revision [\#119](https://github.com/stelligent/mu/pull/119) ([jesseadams](https://github.com/jesseadams))
- Added support for viewing logs from CLI  [\#118](https://github.com/stelligent/mu/pull/118) ([cplee](https://github.com/cplee))
- Service Discovery via Consul [\#117](https://github.com/stelligent/mu/pull/117) ([cplee](https://github.com/cplee))
- Default service name to git repo name [\#113](https://github.com/stelligent/mu/pull/113) ([jesseadams](https://github.com/jesseadams))
- Retrieve git revision from CodePipeline [\#109](https://github.com/stelligent/mu/pull/109) ([jesseadams](https://github.com/jesseadams))

## [v0.1.9](https://github.com/stelligent/mu/tree/v0.1.9) (2017-03-17)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.8...v0.1.9)

**Merged pull requests:**

- CodeCommit support [\#111](https://github.com/stelligent/mu/pull/111) ([cplee](https://github.com/cplee))
- Add support for CodeCommit [\#110](https://github.com/stelligent/mu/pull/110) ([cplee](https://github.com/cplee))
- Updating building from source instructions [\#108](https://github.com/stelligent/mu/pull/108) ([jesseadams](https://github.com/jesseadams))

## [v0.1.8](https://github.com/stelligent/mu/tree/v0.1.8) (2017-03-08)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.7...v0.1.8)

**Merged pull requests:**

- Release 0.1.8 [\#106](https://github.com/stelligent/mu/pull/106) ([cplee](https://github.com/cplee))
- Pipeline managing environments [\#103](https://github.com/stelligent/mu/pull/103) ([cplee](https://github.com/cplee))
- Override CFN resources [\#98](https://github.com/stelligent/mu/pull/98) ([cplee](https://github.com/cplee))
- Execute tests from CodePipeline [\#96](https://github.com/stelligent/mu/pull/96) ([cplee](https://github.com/cplee))
- Add json formatter for env show.  Update pipeline to run tests. [\#95](https://github.com/stelligent/mu/pull/95) ([cplee](https://github.com/cplee))
- Environment variables [\#94](https://github.com/stelligent/mu/pull/94) ([cplee](https://github.com/cplee))
- CloudWatch logs aggregation [\#93](https://github.com/stelligent/mu/pull/93) ([cplee](https://github.com/cplee))

## [v0.1.7](https://github.com/stelligent/mu/tree/v0.1.7) (2017-02-07)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.6...v0.1.7)

**Merged pull requests:**

- v0.1.7 [\#92](https://github.com/stelligent/mu/pull/92) ([cplee](https://github.com/cplee))
- Config and autogeneration of ELB rule priority [\#91](https://github.com/stelligent/mu/pull/91) ([cplee](https://github.com/cplee))
- Add support for HTTPS and DNS to ELB  [\#90](https://github.com/stelligent/mu/pull/90) ([cplee](https://github.com/cplee))
- VPC Enhancements [\#89](https://github.com/stelligent/mu/pull/89) ([cplee](https://github.com/cplee))
- Region and profile [\#75](https://github.com/stelligent/mu/pull/75) ([cplee](https://github.com/cplee))

## [v0.1.6](https://github.com/stelligent/mu/tree/v0.1.6) (2017-02-01)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.5...v0.1.6)

**Merged pull requests:**

- Bug with stack params [\#74](https://github.com/stelligent/mu/pull/74) ([cplee](https://github.com/cplee))

## [v0.1.5](https://github.com/stelligent/mu/tree/v0.1.5) (2017-02-01)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.4...v0.1.5)

**Closed issues:**

- Create CONTRIBUTING.md file [\#50](https://github.com/stelligent/mu/issues/50)

**Merged pull requests:**

- Pipelines [\#73](https://github.com/stelligent/mu/pull/73) ([cplee](https://github.com/cplee))
- Pipeline workflows [\#72](https://github.com/stelligent/mu/pull/72) ([cplee](https://github.com/cplee))
- Create pipeline with CodePipeline and CodeBuild [\#71](https://github.com/stelligent/mu/pull/71) ([cplee](https://github.com/cplee))
- VPC reference from mu.yml [\#70](https://github.com/stelligent/mu/pull/70) ([cplee](https://github.com/cplee))
- Contributing guidelines [\#69](https://github.com/stelligent/mu/pull/69) ([cplee](https://github.com/cplee))
- Pin dependencies with glide [\#67](https://github.com/stelligent/mu/pull/67) ([cplee](https://github.com/cplee))

## [v0.1.4](https://github.com/stelligent/mu/tree/v0.1.4) (2017-01-25)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.3...v0.1.4)

**Merged pull requests:**

- Migrate to CircleCI [\#66](https://github.com/stelligent/mu/pull/66) ([cplee](https://github.com/cplee))
- Migrate to CircleCI [\#65](https://github.com/stelligent/mu/pull/65) ([cplee](https://github.com/cplee))

## [v0.1.3](https://github.com/stelligent/mu/tree/v0.1.3) (2017-01-24)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.2...v0.1.3)

**Merged pull requests:**

- Complete work for supporting deploy/view/undeploy services [\#64](https://github.com/stelligent/mu/pull/64) ([cplee](https://github.com/cplee))
- Handle viewing services and undeploying [\#59](https://github.com/stelligent/mu/pull/59) ([cplee](https://github.com/cplee))
- Complete service view [\#58](https://github.com/stelligent/mu/pull/58) ([cplee](https://github.com/cplee))
- Build/Push/Deploy service [\#57](https://github.com/stelligent/mu/pull/57) ([cplee](https://github.com/cplee))

## [v0.1.2](https://github.com/stelligent/mu/tree/v0.1.2) (2017-01-17)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.1...v0.1.2)

**Merged pull requests:**

- Release 0.1.2 [\#56](https://github.com/stelligent/mu/pull/56) ([cplee](https://github.com/cplee))
- Terminate environments [\#55](https://github.com/stelligent/mu/pull/55) ([cplee](https://github.com/cplee))
- Show environment details [\#54](https://github.com/stelligent/mu/pull/54) ([cplee](https://github.com/cplee))
- Add support to list environments [\#53](https://github.com/stelligent/mu/pull/53) ([cplee](https://github.com/cplee))
- Unit testability for issue 18 [\#52](https://github.com/stelligent/mu/pull/52) ([cplee](https://github.com/cplee))
- Parse mu objects [\#51](https://github.com/stelligent/mu/pull/51) ([cplee](https://github.com/cplee))

## [v0.1.1](https://github.com/stelligent/mu/tree/v0.1.1) (2017-01-05)
[Full Changelog](https://github.com/stelligent/mu/compare/v0.1.0...v0.1.1)

**Implemented enhancements:**

- Create initial CI job [\#8](https://github.com/stelligent/mu/issues/8)

**Merged pull requests:**

- Fix issue with install script picking up from develop branch [\#49](https://github.com/stelligent/mu/pull/49) ([cplee](https://github.com/cplee))

## [v0.1.0](https://github.com/stelligent/mu/tree/v0.1.0) (2017-01-05)
**Merged pull requests:**

- makefile typo [\#16](https://github.com/stelligent/mu/pull/16) ([cplee](https://github.com/cplee))
- Release version 0.1.0 [\#15](https://github.com/stelligent/mu/pull/15) ([cplee](https://github.com/cplee))
- update install instructions and created install script [\#14](https://github.com/stelligent/mu/pull/14) ([cplee](https://github.com/cplee))
- Issue 8 [\#13](https://github.com/stelligent/mu/pull/13) ([cplee](https://github.com/cplee))
- setup git credentials to deploy [\#12](https://github.com/stelligent/mu/pull/12) ([cplee](https://github.com/cplee))
- handle situation where version doesn't already exist [\#11](https://github.com/stelligent/mu/pull/11) ([cplee](https://github.com/cplee))
- Create Makefile and TravisCI for release management [\#10](https://github.com/stelligent/mu/pull/10) ([cplee](https://github.com/cplee))
- layout initial CLI [\#7](https://github.com/stelligent/mu/pull/7) ([cplee](https://github.com/cplee))
- CLI commands [\#6](https://github.com/stelligent/mu/pull/6) ([cplee](https://github.com/cplee))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*