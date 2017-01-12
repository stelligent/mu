package resources

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"strings"
)

// EnvironmentManager defines the env that will be created
type EnvironmentManager interface {
	UpsertEnvironment(environmentName string) error
}

// NewEnvironmentManager will construct a manager for environments
func NewEnvironmentManager(ctx *common.Context) EnvironmentManager {
	environmentManager := new(environmentManagerImpl)
	environmentManager.context = ctx
	return environmentManager
}

// UpsertEnvironment will create a new stack instance and write the template for the stack
func (environmentManager *environmentManagerImpl) UpsertEnvironment(environmentName string) error {
	env, err := environmentManager.getEnvironment(environmentName)
	if err != nil {
		return err
	}

	if env.VpcTarget.VpcID == "" {
		// no target VPC, we manage it
		err = environmentManager.upsertVpc(env)
		if err != nil {
			return err
		}
	}

	err = environmentManager.upsertEcsCluster(env)
	if err != nil {
		return err
	}

	return nil
}

type environmentManagerImpl struct {
	context *common.Context
}

func (environmentManager *environmentManagerImpl) getEnvironment(environmentName string) (*common.Environment, error) {
	ctx := environmentManager.context
	for _, e := range ctx.Config.Environments {
		if strings.EqualFold(environmentName, e.Name) {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("Unable to find environment named '%s' in mu.yml", environmentName)
}

func buildVpcStackName(env *common.Environment) string {
	return fmt.Sprintf("mu-vpc-%s", env.Name)
}
func buildEnvironmentStackName(env *common.Environment) string {
	return fmt.Sprintf("mu-env-%s", env.Name)
}

func (environmentManager *environmentManagerImpl) upsertVpc(env *common.Environment) error {
	// generate the CFN template
	stack := common.NewStack(buildVpcStackName(env))

	err := stack.WriteTemplate("vpc-template.yml", env)
	if err != nil {
		return err
	}

	// create/update the stack
	fmt.Printf("upserting VPC environment:%s stack:%s path:%s\n", env.Name, stack.Name, stack.TemplatePath)
	err = stack.UpsertStack(environmentManager.context.CloudFormation)
	if err != nil {
		return err
	}

	return nil
}

func (environmentManager *environmentManagerImpl) upsertEcsCluster(env *common.Environment) error {
	// generate the CFN template
	stack := common.NewStack(buildEnvironmentStackName(env))

	if env.VpcTarget.VpcID == "" {
		// apply default parameters since we manage the VPC
		vpcStackName := buildVpcStackName(env)
		stack.WithParameter("VpcId", fmt.Sprintf("%s-VpcId", vpcStackName))
		stack.WithParameter("PublicSubnetAZ1Id", fmt.Sprintf("%s-PublicSubnetAZ1Id", vpcStackName))
		stack.WithParameter("PublicSubnetAZ2Id", fmt.Sprintf("%s-PublicSubnetAZ2Id", vpcStackName))
		stack.WithParameter("PublicSubnetAZ3Id", fmt.Sprintf("%s-PublicSubnetAZ3Id", vpcStackName))
	} else {
		// target VPC referenced from config
		stack.WithParameter("VpcId", env.VpcTarget.VpcID)
		for index, subnet := range env.VpcTarget.PublicSubnetIds {
			stack.WithParameter(fmt.Sprintf("PublicSubnetAZ%dId", index+1), subnet)
		}
	}

	err := stack.WriteTemplate("environment-template.yml", env)
	if err != nil {
		return err
	}

	// create/update the stack
	fmt.Printf("upserting environment:%s stack:%s path:%s\n", env.Name, stack.Name, stack.TemplatePath)
	err = stack.UpsertStack(environmentManager.context.CloudFormation)
	if err != nil {
		return err
	}

	return nil
}
