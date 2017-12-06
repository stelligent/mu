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
		Action: func(c *cli.Context) error {
			forceFlag := c.Args().Get(1)
			ctx.ParamManager.SetParam("forceFlag", forceFlag)
			workflow := workflows.NewPurge(ctx, os.Stdout)
			return workflow()
		},
	}
	return cmd
}
