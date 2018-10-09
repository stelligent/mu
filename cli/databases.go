package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
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
			*newDatabaseGetPasswordCommand(ctx),
			*newDatabaseSetPasswordCommand(ctx),
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

func newDatabaseGetPasswordCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "get-password",
		Aliases:   []string{"gp"},
		Usage:     "get-password",
		ArgsUsage: "<environment> [<service>]",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().Get(0)
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "database")
				return errors.New("environment must be provided")
			}
			serviceName := c.Args().Get(1)
			workflow := workflows.DatabaseGetPassword(ctx, environmentName, serviceName)
			return workflow()
		},
	}

	return cmd
}

func newDatabaseSetPasswordCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "set-password",
		Aliases:   []string{"sp"},
		Usage:     "set-password",
		ArgsUsage: "<environment> [<service>]",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().Get(0)
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "database")
				return errors.New("environment must be provided")
			}
			cliExtension := new(common.CliAdditions)
			newPassword, err := cliExtension.GetPasswdPrompt("  Database password: ")
			if err != nil {
				fmt.Println("")
			}
			serviceName := c.Args().Get(1)
			workflow := workflows.DatabaseSetPassword(ctx, environmentName, serviceName, newPassword)
			return workflow()
		},
	}

	return cmd
}
