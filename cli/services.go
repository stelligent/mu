package cli

import (
	"errors"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func newServicesCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    common.SvcCmd,
		Aliases: []string{common.SvcAlias},
		Usage:   common.SvcUsage,
		Subcommands: []cli.Command{
			*newServicesShowCommand(ctx),
			*newServicesPushCommand(ctx),
			*newServicesDeployCommand(ctx),
			*newServicesUndeployCommand(ctx),
			*newServicesLogsCommand(ctx),
			*newServicesExecuteCommand(ctx),
		},
	}

	return cmd
}

func newServicesShowCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      common.ShowCmd,
		Usage:     common.SvcShowUsage,
		ArgsUsage: common.SvcShowUsage,
		Action: func(c *cli.Context) error {
			service := c.Args().First()
			workflow := workflows.NewServiceViewer(ctx, service, ctx.DockerOut)
			return workflow()
		},
	}

	return cmd
}

func newServicesPushCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  common.PushCmd,
		Usage: common.SvcPushCmdUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  common.TagFlagName,
				Usage: common.SvcPushTagFlagUsage,
			},
		},
		Action: func(c *cli.Context) error {
			tag := c.String(common.Tag)
			workflow := workflows.NewServicePusher(ctx, tag, ctx.DockerOut)
			return workflow()
		},
	}

	return cmd
}

func newServicesDeployCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      common.DeployCmd,
		Usage:     common.SvcDeployCmdUsage,
		ArgsUsage: common.EnvArgUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  common.TagFlagName,
				Usage: common.SvcDeployTagFlagUsage,
			},
		},
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == common.Zero {
				cli.ShowCommandHelp(c, common.DeployCmd)
				return errors.New(common.NoEnvValidation)
			}
			tag := c.String(common.Tag)
			workflow := workflows.NewServiceDeployer(ctx, environmentName, tag)
			return workflow()
		},
	}

	return cmd
}

func newServicesUndeployCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      common.UndeployCmd,
		Usage:     common.SvcUndeployCmdUsage,
		ArgsUsage: common.SvcUndeployArgsUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == common.Zero {
				cli.ShowCommandHelp(c, common.UndeployCmd)
				return errors.New(common.NoEnvValidation)
			}
			serviceName := c.Args().Get(common.SvcUndeploySvcFlagIndex)
			workflow := workflows.NewServiceUndeployer(ctx, serviceName, environmentName)
			return workflow()
		},
	}

	return cmd
}

func newServicesLogsCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  common.LogsCmd,
		Usage: common.SvcLogUsage,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  common.ServiceFlag,
				Usage: common.SvcLogServiceFlagUsage,
			},
			cli.BoolFlag{
				Name:  common.FollowFlag,
				Usage: common.FollowUsage,
			},
			cli.DurationFlag{
				Name:  common.SearchDurationFlag,
				Usage: common.SearchDurationUsage,
				Value: common.DefaultLogDurationValue,
			},
		},
		ArgsUsage: common.SvcLogArgUsage,
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == common.Zero {
				cli.ShowCommandHelp(c, common.LogsCmd)
				return errors.New(common.NoEnvValidation)
			}
			serviceName := c.String(common.SvcCmd)

			workflow := workflows.NewServiceLogViewer(ctx, c.Duration(common.SearchDuration), c.Bool(common.Follow), environmentName, serviceName, os.Stdout, strings.Join(c.Args().Tail(), common.Space))
			return workflow()
		},
	}

	return cmd
}

func validateExecuteArguments(ctx *cli.Context) error {
	environmentName := ctx.Args().First()
	argLength := len(ctx.Args())

	if argLength == common.Zero || len(strings.TrimSpace(environmentName)) == common.Zero {
		cli.ShowCommandHelp(ctx, common.ExeCmd)
		return errors.New(common.NoEnvValidation)
	}
	if argLength == common.ExeArgsCmdIndex {
		cli.ShowCommandHelp(ctx, common.ExeCmd)
		return errors.New(common.NoCmdValidation)
	}
	if len(strings.TrimSpace(ctx.Args().Get(common.ExeArgsCmdIndex))) == common.Zero {
		cli.ShowCommandHelp(ctx, common.ExeCmd)
		return errors.New(common.EmptyCmdValidation)
	}
	return nil
}

func newServicesExecuteCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      common.ExeCmd,
		Usage:     common.ExeUsage,
		ArgsUsage: common.ExeArgs,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  common.ServiceFlag,
				Usage: common.SvcExeServiceFlagUsage,
			},
			cli.StringFlag{
				Name:   common.TaskFlag,
				Usage:  common.SvcExeTaskFlagUsage,
				Hidden: common.TaskFlagVisible,
			},
			cli.StringFlag{
				Name:   common.ClusterFlag,
				Usage:  common.SvcExeClusterFlagUsage,
				Hidden: common.ClusterFlagVisible,
			},
		},
		Action: func(c *cli.Context) error {
			task, err := newTask(c)
			if err != nil {
				return err
			}

			workflow := workflows.NewServiceExecutor(ctx, *task)
			return workflow()
		},
	}
	return cmd
}

func newTask(c *cli.Context) (*common.Task, error) {
	err := validateExecuteArguments(c)
	if err != nil {
		return nil, err
	}
	environmentName := c.Args().First()
	command := strings.Join(c.Args()[common.ExeArgsCmdIndex:], common.Space)
	return &common.Task{
		Environment:    environmentName,
		Command:        command,
		Service:        c.String(common.SvcCmd),
		TaskDefinition: c.String(common.TaskFlagName),
		Cluster:        c.String(common.ClusterFlagName),
	}, nil
}
