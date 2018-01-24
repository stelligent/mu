package cli

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
)

func newEnvironmentsCommand(ctx *common.Context) *cli.Command {

	cmd := &cli.Command{
		Name:    EnvCmd,
		Aliases: []string{EnvAlias},
		Usage:   EnvUsage,
		Subcommands: []cli.Command{
			*newEnvironmentsListCommand(ctx),
			*newEnvironmentsShowCommand(ctx),
			*newEnvironmentsUpsertCommand(ctx),
			*newEnvironmentsTerminateCommand(ctx),
			*newEnvironmentsLogsCommand(ctx),
		},
	}

	return cmd
}

func newEnvironmentsUpsertCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      UpsertCmd,
		Aliases:   []string{UpsertAlias},
		Usage:     UpsertUsage,
		ArgsUsage: EnvArgUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == Zero {
				cli.ShowCommandHelp(c, UpsertCmd)
				return errors.New(NoEnvValidation)
			}

			workflow := workflows.NewEnvironmentUpserter(ctx, environmentName)
			return workflow()
		},
	}

	return cmd
}

func newEnvironmentsListCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    ListCmd,
		Aliases: []string{ListAlias},
		Usage:   ListUsage,
		Action: func(c *cli.Context) error {
			workflow := workflows.NewEnvironmentLister(ctx, os.Stdout)
			return workflow()
		},
	}

	return cmd
}

func newEnvironmentsShowCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  ShowCmd,
		Usage: ShowCmdUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  FormatFlag,
				Usage: FormatFlagUsage,
				Value: FormatFlagDefault,
			},
			cli.BoolFlag{
				Name:  "tasks, t",
				Usage: "show task detail",
			},
			cli.BoolFlag{
				Name:  "watch, w",
				Usage: "watch results",
			},
		},
		ArgsUsage: EnvArgUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == Zero {
				cli.ShowCommandHelp(c, ShowCmd)
				return errors.New(NoEnvValidation)
			}

			viewTasks := c.Bool("tasks")
			watch := c.Bool("watch")
			workflow := workflows.NewEnvironmentViewer(ctx, c.String(Format), environmentName, viewTasks, os.Stdout)
			for true {
				if watch {
					print("\033[H\033[2J")
				}

				err := workflow()
				if err != nil {
					return err
				} else if watch {
					time.Sleep(10 * time.Second)
				} else {
					break
				}
			}
			return nil
		},
	}

	return cmd
}

func newEnvironmentsTerminateCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      TerminateCmd,
		Aliases:   []string{TerminateAlias},
		Usage:     TerminateUsage,
		ArgsUsage: EnvArgUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == Zero {
				cli.ShowCommandHelp(c, TerminateCmd)
				return errors.New(NoEnvValidation)
			}
			workflow := workflows.NewEnvironmentTerminator(ctx, environmentName)
			return workflow()
		},
	}

	return cmd
}

func newEnvironmentsLogsCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  LogsCmd,
		Usage: LogsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  FollowFlag,
				Usage: FollowUsage,
			},
			cli.DurationFlag{
				Name:  SearchDurationFlag,
				Usage: SearchDurationUsage,
				Value: DefaultLogDurationValue,
			},
		},
		ArgsUsage: LogsArgs,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == Zero {
				cli.ShowCommandHelp(c, LogsCmd)
				return errors.New(NoEnvValidation)
			}

			workflow := workflows.NewEnvironmentLogViewer(ctx, c.Duration(SearchDuration), c.Bool(Follow), environmentName, os.Stdout, strings.Join(c.Args().Tail(), Space))
			return workflow()
		},
	}

	return cmd
}
