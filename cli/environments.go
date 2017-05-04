package cli

import (
	"errors"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
	"os"
	"strings"
	"time"
)

func newEnvironmentsCommand(ctx *common.Context) *cli.Command {

	cmd := &cli.Command{
		Name:    common.EnvCmd,
		Aliases: []string{common.EnvAlias},
		Usage:   common.EnvUsage,
		Subcommands: []cli.Command{
			*newEnvironmentsListCommand(ctx),
			*newEnvironmentsShowCommand(ctx),
			*newEnvironmentsUpsertCommand(ctx),
			*newEnvironmentsTerminateCommand(ctx),
			*newEnvironmentsLogsCommand(ctx),
			*newEnvironmentsExecuteCommand(ctx),
		},
	}

	return cmd
}

func newEnvironmentsUpsertCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      common.UpsertCmd,
		Aliases:   []string{common.UpsertAlias},
		Usage:     common.UpsertUsage,
		ArgsUsage: common.EnvArgUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, common.UpsertCmd)
				return errors.New(common.NoEnvValidation)
			}

			workflow := workflows.NewEnvironmentUpserter(ctx, environmentName)
			return workflow()
		},
	}

	return cmd
}

func newEnvironmentsListCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    common.ListCmd,
		Aliases: []string{common.ListAlias},
		Usage:   common.ListUsage,
		Action: func(c *cli.Context) error {
			workflow := workflows.NewEnvironmentLister(ctx, os.Stdout)
			return workflow()
		},
	}

	return cmd
}

func newEnvironmentsShowCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  common.ShowCmd,
		Usage: common.ShowCmdUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  common.FormatFlag,
				Usage: common.FormatFlagUsage,
				Value: common.FormatFlagDefault,
			},
		},
		ArgsUsage: common.EnvArgUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, common.ShowCmd)
				return errors.New(common.NoEnvValidation)
			}
			workflow := workflows.NewEnvironmentViewer(ctx, c.String(common.Format), environmentName, os.Stdout)
			return workflow()
		},
	}

	return cmd
}

func newEnvironmentsTerminateCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      common.TerminateCmd,
		Aliases:   []string{common.TerminateAlias},
		Usage:     common.TerminateUsage,
		ArgsUsage: common.EnvArgUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, common.TerminateCmd)
				return errors.New(common.NoEnvValidation)
			}
			workflow := workflows.NewEnvironmentTerminator(ctx, environmentName)
			return workflow()
		},
	}

	return cmd
}

func validateExecuteArguments(ctx *cli.Context) error {
	environmentName := ctx.Args().First()
	argLength := len(ctx.Args())

	if argLength == 0 || len(strings.TrimSpace(environmentName)) == 0 {
		cli.ShowCommandHelp(ctx, common.ExeCmd)
		return errors.New(common.NoEnvValidation)
	}

	if argLength == 1 {
		cli.ShowCommandHelp(ctx, common.ExeCmd)
		return errors.New(common.NoCmdValidation)
	}

	if len(strings.TrimSpace(ctx.Args().Get(1))) == 0 {
		cli.ShowCommandHelp(ctx, common.ExeCmd)
		return errors.New(common.EmptyCmdValidation)
	}
	return nil
}

func newEnvironmentsExecuteCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      common.ExeCmd,
		Usage:     common.ExeUsage,
		ArgsUsage: common.ExeArgs,
		Action: func(c *cli.Context) error {
			err := validateExecuteArguments(c)
			if err != nil {
				return err
			}
			environmentName := c.Args().First()
			workflow := workflows.NewEnvironmentExecutor(ctx, environmentName, strings.Join(c.Args().Tail(), common.Space))
			return workflow()
		},
	}
	return cmd
}

func newEnvironmentsLogsCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  common.LogsCmd,
		Usage: common.LogsUsage,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  common.FollowFlag,
				Usage: common.FollowUsage,
			},
			cli.DurationFlag{
				Name:  common.SearchDurationFlag,
				Usage: common.SearchDurationUsage,
				Value: 1 * time.Minute,
			},
		},
		ArgsUsage: common.LogsArgs,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, common.LogsCmd)
				return errors.New(common.NoEnvValidation)
			}

			workflow := workflows.NewEnvironmentLogViewer(ctx, c.Duration(common.SearchDuration), c.Bool(common.Follow), environmentName, os.Stdout, strings.Join(c.Args().Tail(), common.Space))
			return workflow()
		},
	}

	return cmd
}
