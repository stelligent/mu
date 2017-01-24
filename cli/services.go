package cli

import (
	"errors"
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
	"os"
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
			*newServicesSetenvCommand(ctx),
			*newServicesUndeployCommand(ctx),
		},
	}

	return cmd
}

func newServicesShowCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "show",
		Usage: "show service details",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: "service to show",
			},
		},
		Action: func(c *cli.Context) error {
			service := c.String("service")
			workflow := workflows.NewServiceViewer(ctx, service, os.Stdout)
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
			workflow := workflows.NewServicePusher(ctx, tag, os.Stdout)
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
				Usage: "tag to push",
			},
		},
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if len(environmentName) == 0 {
				cli.ShowCommandHelp(c, "terminate")
				return errors.New("environment must be provided")
			}
			tag := c.String("tag")
			workflow := workflows.NewServiceDeployer(ctx, environmentName, tag)
			return workflow()
		},
	}

	return cmd
}

func newServicesSetenvCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "setenv",
		Usage:     "set environment variable",
		ArgsUsage: "<environment> <key1>=<value1>...",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: "service to deploy",
			},
		},
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			serviceName := c.String("service")
			keyvals := c.Args().Tail()
			fmt.Printf("setenv service: %s to environment: %s with vals: %s\n", serviceName, environmentName, keyvals)
			return nil
		},
	}

	return cmd
}

func newServicesUndeployCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "undeploy",
		Usage:     "undeploy service from environment",
		ArgsUsage: "<environment>",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: "service to undeploy",
			},
		},
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			serviceName := c.String("service")
			fmt.Printf("undeploying service: %s to environment: %s\n", serviceName, environmentName)
			return nil
		},
	}

	return cmd
}
