package environments

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

func newUpsertCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "upsert",
		Aliases:   []string{"up"},
		Usage:     "create/update an environment",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if(len(environmentName) == 0) {
				cli.ShowCommandHelp(c, "upsert")
				return fmt.Errorf("environment must be provided")
			}
			return runUpsert(config, environmentName)
		},
	}

	return cmd
}

func runUpsert(config *common.Config, environmentName string) error {
	// get the environment from config by name
	environment, err := config.GetEnvironment(environmentName)
	if err != nil {
		return err
	}

	// generate the CFN template
	stack, err := environment.NewStack()
	if err != nil {
		return err
	}

	// determine if stack exists

	// create/update the stack

	// wait for stack to be updated

	fmt.Printf("upserting environment:%s stack:%s path:%s\n",environment.Name, stack.Name, stack.TemplatePath)

	return nil
}



