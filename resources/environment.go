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
func NewEnvironmentManager(config *common.Config) (EnvironmentManager) {
	ctx := new(environmentManagerContext)
	ctx.config = config
	return ctx
}

type environmentManagerContext struct {
	config *common.Config
}

// getEnvironment loads the environment by name from the config
func (ctx *environmentManagerContext) getEnvironment(environmentName string) (*common.Environment, error) {

	for _, e := range ctx.config.Environments {
		if(strings.EqualFold(environmentName, e.Name)) {
			return &e, nil
		}
	}

	return nil, fmt.Errorf("Unable to find environment named '%s' in mu.yml",environmentName)
}

// getRegion determines the region to use
func (ctx *environmentManagerContext) getRegion() (string) {
	return ctx.config.Region
}

// UpsertRegion will create a new stack instance and write the template for the stack
func (ctx *environmentManagerContext) UpsertEnvironment(environmentName string) (error) {
	err := ctx.upsertVpc(environmentName)
	if err != nil {
		return err
	}

	err = ctx.upsertEcsCluster(environmentName)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *environmentManagerContext) upsertVpc(environmentName string) (error) {
	env, err := ctx.getEnvironment(environmentName)
	if err != nil {
		return err
	}

	stackName := fmt.Sprintf("mu-vpc-%s", env.Name)
	// generate the CFN template
	stack := common.NewStack(stackName, ctx.getRegion())

	err = stack.WriteTemplate("vpc-template.yml", env)
	if err != nil {
		return err
	}

	// create/update the stack
	fmt.Printf("upserting VPC environment:%s stack:%s path:%s\n",env.Name, stack.Name, stack.TemplatePath)
	err = stack.UpsertStack()
	if err != nil {
		return err
	}

	return nil
}

func (ctx *environmentManagerContext) upsertEcsCluster(environmentName string) (error) {
	env, err := ctx.getEnvironment(environmentName)
	if err != nil {
		return err
	}

	stackName := fmt.Sprintf("mu-env-%s", env.Name)
	// generate the CFN template
	stack := common.NewStack(stackName, ctx.getRegion())

	err = stack.WriteTemplate("environment-template.yml", env)
	if err != nil {
		return err
	}

	// create/update the stack
	fmt.Printf("upserting environment:%s stack:%s path:%s\n",env.Name, stack.Name, stack.TemplatePath)
	err = stack.UpsertStack()
	if err != nil {
		return err
	}

	return nil
}


