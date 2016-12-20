package main

import (
    "os"
    "github.com/urfave/cli"
    "github.com/stelligent/mu/environments"
    "github.com/stelligent/mu/services"
    "github.com/stelligent/mu/pipelines"
)

var version string

func main() {
    app := cli.NewApp()
    app.Name = "mu"
    app.Usage = "Microservice Platform on AWS"
    app.Version = version
    app.EnableBashCompletion = true

    app.Commands = []cli.Command{
        {
            Name: "environment",
            Aliases: []string{"env"},
            Usage: "options for managing environments",
            Subcommands: []cli.Command{
                {
                    Name: "list",
                    Aliases: []string{"ls"},
                    Usage: "list environments",
                    Action: func(c *cli.Context) error {
                        environments.List()
                        return nil
                    },
                },
                {
                    Name: "show",
                    Usage: "show environment details",
                    ArgsUsage: "<environment>",
                    Action: func(c *cli.Context) error {
                        environments.Show(c.Args().First())
                        return nil
                    },
                },
                {
                    Name: "upsert",
                    Aliases: []string{"up"},
                    Usage: "create/update an environment",
                    ArgsUsage: "<environment>",
                    Action: func(c *cli.Context) error {
                        environments.Upsert(c.Args().First())
                        return nil
                    },
                },
                {
                    Name: "terminate",
                    Aliases: []string{"term"},
                    Usage: "terminate an environment",
                    ArgsUsage: "<environment>",
                    Action: func(c *cli.Context) error {
                        environments.Terminate(c.Args().First())
                        return nil
                    },
                },
            },
        },
        {
            Name: "service",
            Aliases: []string{"svc"},
            Usage: "options for managing services",
            Subcommands: []cli.Command{
                {
                    Name: "show",
                    Usage: "show service details",
                    Flags: []cli.Flag {
                        cli.StringFlag{
                            Name: "service, s",
                            Usage: "service to show",
                        },
                    },
                    Action: func(c *cli.Context) error {
                        services.Show(c.String("service"))
                        return nil
                    },
                },
                {
                    Name: "deploy",
                    Usage: "deploy service to environment",
                    ArgsUsage: "<environment>",
                    Flags: []cli.Flag {
                        cli.StringFlag{
                            Name: "service, s",
                            Usage: "service to deploy",
                        },
                    },
                    Action: func(c *cli.Context) error {
                        services.Deploy(c.Args().First(), c.String("service"))
                        return nil
                    },
                },
                {
                    Name: "setenv",
                    Usage: "set environment variable",
                    ArgsUsage: "<environment> <key1>=<value1>...",
                    Flags: []cli.Flag {
                        cli.StringFlag{
                            Name: "service, s",
                            Usage: "service to deploy",
                        },
                    },
                    Action: func(c *cli.Context) error {
                        services.Setenv(c.Args().First(), c.String("service"), c.Args().Tail())
                        return nil
                    },
                },
                {
                    Name: "undeploy",
                    Usage: "undeploy service from environment",
                    ArgsUsage: "<environment>",
                    Flags: []cli.Flag {
                        cli.StringFlag{
                            Name: "service, s",
                            Usage: "service to undeploy",
                        },
                    },
                    Action: func(c *cli.Context) error {
                        services.Undeploy(c.Args().First(), c.String("service"))
                        return nil
                    },
                },
            },
        },
        {
            Name: "pipeline",
            Usage: "options for managing pipelines",
            Subcommands: []cli.Command{
                {
                    Name: "list",
                    Usage: "list pipelines",
                    Action: func(c *cli.Context) error {
                        pipelines.List()
                        return nil
                    },
                },
                {
                    Name: "show",
                    Usage: "show pipeline details",
                    Flags: []cli.Flag {
                        cli.StringFlag{
                            Name: "service, s",
                            Usage: "service to show",
                        },
                    },
                    Action: func(c *cli.Context) error {
                        pipelines.Show(c.String("service"))
                        return nil
                    },
                },
            },
        },
    }

    app.Run(os.Args)
}

