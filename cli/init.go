package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
)

func newInitCommand(ctx *common.Context) *cli.Command {

	cmd := &cli.Command{
		Name: "init",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "env, e",
				Usage: "Initialize environments as well as service (default: false)",
			},
			cli.IntFlag{
				Name:  "port, P",
				Usage: "Port the application listens on",
				Value: 8080,
			},
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Force overwrite of existing mu.yml (default: false)",
			},
		},
		Usage: "initialize mu.yml file",
		Action: func(c *cli.Context) error {
			workflow := workflows.NewConfigInitializer(ctx, c.Bool("env"), c.Int("port"), c.Bool("force"))
			return workflow()
		},
	}

	return cmd
}
