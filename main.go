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
    config := common.LoadConfig()

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

    app.Run(os.Args)
}

