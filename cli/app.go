package cli

import (
	"github.com/urfave/cli"
	"github.com/stelligent/mu/common"
)

// NewApp creates a new CLI app
func NewApp(version string) *cli.App {
	context := common.NewContext()

	app := cli.NewApp()
	app.Name = "mu"
	app.Usage = "Microservice Platform on AWS"
	app.Version = version
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		*newEnvironmentsCommand(context),
		*newServicesCommand(context),
		*newPipelinesCommand(context),
	}

	app.Before = func(c *cli.Context) error {
		context.InitializeFromFile(c.String("config"))
		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "config, c",
			Usage: "path to config file",
			Value: "mu.yml",
		},
	}

	return app
}

