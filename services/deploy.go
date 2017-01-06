package services

import(
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

func newDeployCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "deploy",
		Usage: "deploy service to environment",
		ArgsUsage: "<environment>",
		Flags: []cli.Flag {
			cli.StringFlag{
				Name: "service, s",
				Usage: "service to deploy",
			},
		},
		Action: func(c *cli.Context) error {
			runDeploy(config, c.Args().First(), c.String("service"))
			return nil
		},
	}

	return cmd
}

func runDeploy(config *common.Config, environment string, service string) {
	fmt.Printf("deploying service: %s to environment: %s\n",service, environment)
}

