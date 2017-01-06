package main

import (
    "os"
    "github.com/urfave/cli"
    "github.com/stelligent/mu/environments"
    "github.com/stelligent/mu/services"
    "github.com/stelligent/mu/pipelines"
    "github.com/stelligent/mu/common"
)

var version string

func main() {
    app := newApp()
    app.Run(os.Args)
}

func newApp() *cli.App {
    config := common.NewConfig()
    app := cli.NewApp()
    app.Name = "mu"
    app.Usage = "Microservice Platform on AWS"
    app.Version = version
    app.EnableBashCompletion = true

    app.Commands = []cli.Command{
        *environments.NewCommand(config),
        *services.NewCommand(config),
        *pipelines.NewCommand(config),
    }

    app.Before = func(c *cli.Context) error {
        common.LoadConfig(config, c.String("config"))
        return nil
    }

    app.Flags = []cli.Flag {
        cli.StringFlag{
            Name: "config, c",
            Usage: "path to config file",
            Value: "mu.yml",
        },
    }

    return app
}

