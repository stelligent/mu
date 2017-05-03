package cli

import (
	"errors"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
	"os"
)

func newDatabasesCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "database",
		Aliases: []string{"db"},
		Usage:   "options for managing databases",
		Subcommands: []cli.Command{
			*newDatabaseListCommand(ctx),
			*newDatabaseUpsertCommand(ctx),
			*newDatabaseTerminateCommand(ctx),
		},
	}

	return cmd
}

func newDatabaseListCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list databases",
		Action: func(c *cli.Context) error {
			workflow := workflows.NewDatabaseLister(ctx, os.Stdout)
			return workflow()
		},
	}

	return cmd
}

func newDatabaseTerminateCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "terminate",
		Aliases:   []string{"term"},
		Usage:     "terminate database",
		ArgsUsage: "<environment> [<service>]",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "terminate")
				return errors.New("environment must be provided")
			}
			serviceName := c.Args().Get(1)
			workflow := workflows.NewDatabaseTerminator(ctx, serviceName, environmentName)
			return workflow()
		},
	}

	return cmd
}

func newDatabaseUpsertCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "upsert",
		Aliases:   []string{"up"},
		Usage:     "upsert database",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "deploy")
				return errors.New("environment must be provided")
			}
			workflow := workflows.NewDatabaseUpserter(ctx, environmentName)
			return workflow()
		},
	}

	return cmd
}
