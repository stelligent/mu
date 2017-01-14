package cli

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

func newPipelinesCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "pipeline",
		Usage: "options for managing pipelines",
		Subcommands: []cli.Command{
			*newPipelinesListCommand(ctx),
			*newPipelinesShowCommand(ctx),
		},
	}

	return cmd
}

func newPipelinesListCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list pipelines",
		Action: func(c *cli.Context) error {
			fmt.Println("listing pipelines")
			return nil
		},
	}

	return cmd
}

func newPipelinesShowCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "show",
		Usage: "show pipeline details",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: "service to show",
			},
		},
		Action: func(c *cli.Context) error {
			service := c.String("service")
			fmt.Printf("showing pipeline: %s\n", service)
			return nil
		},
	}

	return cmd
}
