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
			paramName := "yes"
			suppressConfirmation := c.Bool(paramName)
			if suppressConfirmation {
				ctx.ParamManager.SetParam("suppressConfirmation", "yes")
			} else {
				ctx.ParamManager.SetParam("suppressConfirmation", "no")
			}
			workflow := workflows.NewPurge(ctx, os.Stdout)
			return workflow()
		},
	}
	return cmd
}
