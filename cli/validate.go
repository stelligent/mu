package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

func newValidateCommand(ctx *common.Context) *cli.Command {

	cmd := &cli.Command{
		Name:  "validate",
		Usage: "validate mu config",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
	return cmd
}
