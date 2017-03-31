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
		Name:    "service",
		Aliases: []string{"svc"},
		Usage:   "options for managing services",
		Subcommands: []cli.Command{
			*newServicesShowCommand(ctx),
			*newServicesPushCommand(ctx),
			*newServicesDeployCommand(ctx),
			*newServicesUndeployCommand(ctx),
			*newServicesLogsCommand(ctx),
		},
	}

	return cmd
}

func newServicesShowCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "show",
		Usage:     "show service details",
		ArgsUsage: "[<service>]",
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
		Name:  "push",
		Usage: "push service to repository",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "tag, t",
				Usage: "tag to push",
			},
		},
		Action: func(c *cli.Context) error {
			tag := c.String("tag")
			workflow := workflows.NewServicePusher(ctx, tag, ctx.DockerOut)
			return workflow()
		},
	}

	return cmd
}

func newServicesDeployCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "deploy",
		Usage:     "deploy service to environment",
		ArgsUsage: "<environment>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "tag, t",
				Usage: "tag to deploy",
			},
		},
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "deploy")
				return errors.New("environment must be provided")
			}
			tag := c.String("tag")
			workflow := workflows.NewServiceDeployer(ctx, environmentName, tag)
			return workflow()
		},
	}

	return cmd
}

func newServicesUndeployCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "undeploy",
		Usage:     "undeploy service from environment",
		ArgsUsage: "<environment> [<service>]",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "undeploy")
				return errors.New("environment must be provided")
			}
			serviceName := c.Args().Get(1)
			workflow := workflows.NewServiceUndeployer(ctx, serviceName, environmentName)
			return workflow()
		},
	}

	return cmd
}
func newServicesLogsCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "logs",
		Usage: "show service logs",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: "service name to view logs for",
			},
			cli.BoolFlag{
				Name:  "follow, f",
				Usage: "follow logs for latest changes",
			},
		},
		ArgsUsage: "<environment> [<filter>...]",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "logs")
				return errors.New("environment must be provided")
			}
			serviceName := c.String("service")

			workflow := workflows.NewServiceLogViewer(ctx, c.Bool("follow"), environmentName, serviceName, os.Stdout, strings.Join(c.Args().Tail(), " "))
			return workflow()
		},
	}

	return cmd
}
