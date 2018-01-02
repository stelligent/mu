package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
	"os"
)

func newPurgeCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "purge",
		Aliases: []string{"nuke"},
		Usage:   "purge",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "yes, y",
				Usage: "suppresses (Y/N) confirmation prompt",
			},
			cli.StringFlag{
				Name:  "namespace, n",
				Usage: "specify a namespace to filter",
			},
		},
		Action: func(c *cli.Context) error {
			suppressConfirmation := c.Bool("yes")
			workflow := workflows.NewPurge(ctx, suppressConfirmation, os.Stdout)
			return workflow()
		},
	}
	return cmd
}
