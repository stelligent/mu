package services

import(
	"fmt"
	"github.com/urfave/cli"
	"github.com/stelligent/mu/common"
)

func newUndeployCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "undeploy",
		Usage: "undeploy service from environment",
		ArgsUsage: "<environment>",
		Flags: []cli.Flag {
			cli.StringFlag{
				Name: "service, s",
				Usage: "service to undeploy",
			},
		},
		Action: func(c *cli.Context) error {
			runUndeploy(config, c.Args().First(), c.String("service"))
			return nil
		},
	}

	return cmd
}

func runUndeploy(config *common.Config, environment string, service string) {
	fmt.Printf("undeploying service: %s to environment: %s\n",service, environment)
}
