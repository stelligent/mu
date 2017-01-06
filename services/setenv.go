package services

import(
	"fmt"
	"github.com/urfave/cli"
	"github.com/stelligent/mu/common"
)

func newSetenvCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "setenv",
		Usage: "set environment variable",
		ArgsUsage: "<environment> <key1>=<value1>...",
		Flags: []cli.Flag {
			cli.StringFlag{
				Name: "service, s",
				Usage: "service to deploy",
			},
		},
		Action: func(c *cli.Context) error {
			runSetenv(config, c.Args().First(), c.String("service"), c.Args().Tail())
			return nil
		},
	}

	return cmd
}

func runSetenv(config *common.Config, environment string, service string, keyvals []string) {
	fmt.Printf("setenv service: %s to environment: %s with vals: %s\n",service, environment, keyvals)
}
