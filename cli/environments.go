package cli

import (
	"errors"
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/resources"
	"github.com/urfave/cli"
)

func newEnvironmentsCommand(ctx *common.Context) *cli.Command {

	cmd := &cli.Command{
		Name:    "environment",
		Aliases: []string{"env"},
		Usage:   "options for managing environments",
		Subcommands: []cli.Command{
			*newEnvironmentsListCommand(ctx),
			*newEnvironmentsShowCommand(ctx),
			*newEnvironmentsUpsertCommand(ctx),
			*newEnvironmentsTerminateCommand(ctx),
		},
	}

	return cmd
}

func newEnvironmentsUpsertCommand(ctx *common.Context) *cli.Command {
	environmentManager := resources.NewEnvironmentManager(ctx)
	cmd := &cli.Command{
		Name:      "upsert",
		Aliases:   []string{"up"},
		Usage:     "create/update an environment",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "upsert")
				return errors.New("environment must be provided")
			}

			return environmentManager.UpsertEnvironment(environmentName)
		},
	}

	return cmd
}

func newEnvironmentsListCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list environments",
		Action: func(c *cli.Context) error {
			fmt.Println("listing environments")
			return nil
		},
	}

	return cmd
}

func newEnvironmentsShowCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "show",
		Usage:     "show environment details",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "show")
				return errors.New("environment must be provided")
			}
			fmt.Printf("showing environment: %s\n", environmentName)
			return nil
		},
	}

	return cmd
}
func newEnvironmentsTerminateCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "terminate",
		Aliases:   []string{"term"},
		Usage:     "terminate an environment",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "terminate")
				return errors.New("environment must be provided")
			}
			fmt.Printf("terminating environment: %s\n", environmentName)
			return nil
		},
	}

	return cmd
}
