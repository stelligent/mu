package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
)

func newCatalogCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "catalog",
		Aliases: []string{"cat"},
		Usage:   "options for managing catalogs",
		Subcommands: []cli.Command{
			*newCatalogUpsertCommand(ctx),
			*newCatalogTerminateCommand(ctx),
		},
	}
	return cmd
}

func newCatalogTerminateCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "terminate",
		Aliases: []string{"term"},
		Usage:   "terminate catalog",
		Action: func(c *cli.Context) error {
			workflow := workflows.NewCatalogTerminator(ctx)
			return workflow()
		},
	}

	return cmd
}

func newCatalogUpsertCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "upsert",
		Aliases: []string{"up"},
		Usage:   "upsert catalog",
		Action: func(c *cli.Context) error {
			workflow := workflows.NewCatalogUpserter(ctx)
			return workflow()
		},
	}

	return cmd
}
