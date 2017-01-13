package cli

import (
	"bufio"
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
	"os"
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
		// setup logging
		if c.Bool("verbose") {
			common.SetupLogging(2)
		} else if c.Bool("silent") {
			common.SetupLogging(0)
		} else {
			common.SetupLogging(1)

		}

		// load yaml config
		yamlFile, err := os.Open(c.String("config"))
		if err != nil {
			return err
		}
		defer func() {
			yamlFile.Close()
		}()

		// initialize context
		context.Initialize(bufio.NewReader(yamlFile))
		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "path to config file",
			Value: "mu.yml",
		},
		cli.BoolFlag{
			Name:  "silent, s",
			Usage: "silent mode, errors only",
		},
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "increase level of log verbosity",
		},
	}

	return app
}
