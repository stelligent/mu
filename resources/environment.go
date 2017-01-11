package resources

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"strings"
)


// EnvironmentManager defines the env that will be created
type EnvironmentManager interface {
	UpsertEnvironment(environmentName string) (error)
}


// NewEnvironmentManager will construct a manager for environments
func NewEnvironmentManager(ctx *common.Context) (EnvironmentManager) {
	environmentManager := new(environmentManagerImpl)
	environmentManager.context = ctx
	return environmentManager
}

// UpsertEnvironment will create a new stack instance and write the template for the stack
func (environmentManager *environmentManagerImpl) UpsertEnvironment(environmentName string) (error) {
	env, err := environmentManager.getEnvironment(environmentName)
	if err != nil {
		return err
	}

	err = environmentManager.upsertVpc(env)
	if err != nil {
		return err
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
		if(strings.EqualFold(environmentName, e.Name)) {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("Unable to find environment named '%s' in mu.yml",environmentName)
}

func (environmentManager *environmentManagerImpl) upsertVpc(env *common.Environment) (error) {
	cfn := environmentManager.context.CloudFormation
	stackName := fmt.Sprintf("mu-vpc-%s", env.Name)
	// generate the CFN template
	stack := common.NewStack(stackName)

	err := stack.WriteTemplate("vpc-template.yml", env)
	if err != nil {
		return err
	}

	// create/update the stack
	fmt.Printf("upserting VPC environment:%s stack:%s path:%s\n",env.Name, stack.Name, stack.TemplatePath)
	err = stack.UpsertStack(cfn)
	if err != nil {
		return err
	}

	return nil
}

func (environmentManager *environmentManagerImpl) upsertEcsCluster(env *common.Environment) (error) {
	cfn := environmentManager.context.CloudFormation
	stackName := fmt.Sprintf("mu-env-%s", env.Name)
	// generate the CFN template
	stack := common.NewStack(stackName)

	err := stack.WriteTemplate("environment-template.yml", env)
	if err != nil {
		return err
	}

	// create/update the stack
	fmt.Printf("upserting environment:%s stack:%s path:%s\n",env.Name, stack.Name, stack.TemplatePath)
	err = stack.UpsertStack(cfn)
	if err != nil {
		return err
	}

	return nil
}


