package cli

import (
	"fmt"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
)

func newPurgeCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "purge",
		Usage: "purge all resources for a namespace",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "confirm",
				Usage: "confirm deletion of all resources",
			},
		},
		Action: func(c *cli.Context) error {
			confirmation := c.Bool("confirm")
			if !confirmation {
				cliExtension := new(common.CliAdditions)
				confirmation, err := cliExtension.Prompt(fmt.Sprintf("Are you sure you wish to delete all resources in the '%s' namespace?", ctx.Config.Namespace), false)
				if err != nil {
					return err
				}
				if !confirmation {
					return fmt.Errorf("aborted")
				}
			}
			workflow := workflows.NewPurge(ctx)
			return workflow()
		},
	}
	return cmd
}
