package cli

import (
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/provider/aws"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
)

// NewApp creates a new CLI app
func NewApp() *cli.App {
	context := common.NewContext()

	app := cli.NewApp()
	app.Name = "mu"
	app.Usage = "Microservice Platform on AWS"
	app.Version = common.GetVersion()
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		*newInitCommand(context),
		*newEnvironmentsCommand(context),
		*newServicesCommand(context),
		*newPipelinesCommand(context),
		*newDatabasesCommand(context),
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

		// initialize context
		err := context.InitializeContext()
		if err != nil {
			return err
		}

		err = context.InitializeConfigFromFile(c.String("config"))
		if err != nil {
			// ignore errors for init command
			if c.Args().First() != "init" {
				log.Warningf("Unable to load mu config: %v", err)
			}
		}

		// Allow overriding the `DisableIAM` in config via `--disable-iam` or `-I`
		if c.Bool("disable-iam") {
			context.Config.DisableIAM = true
		}

		// TODO: support initializing context from other cloud providers?
		err = aws.InitializeContext(context, c.String("profile"), c.String("region"), c.Bool("dryrun"))
		if err != nil {
			return err
		}

		if c.Bool("silent") {
			context.DockerOut = ioutil.Discard
		} else {
			context.DockerOut = os.Stdout
		}

		// Get the namespace for the stack creation.  This will prefix the stack names
		// The order of precedence is command-line arg, env variable then config file
		nameSpace := c.String("namespace")
		if nameSpace != "" {
			log.Debug("Using namespace \"%s\"", nameSpace)
			context.Config.Namespace = nameSpace
		} else {
			nameSpace = os.Getenv("MU_NAMESPACE")
			if nameSpace != "" {
				log.Debug("Using namespace \"%s\"", nameSpace)
				context.Config.Namespace = nameSpace
			}
		}
		if context.Config.Namespace == "" {
			log.Debug("Using namespace \"mu\"")
			context.Config.Namespace = "mu"
		}

		return nil

	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "path to config file",
			Value: "mu.yml",
		},
		cli.StringFlag{
			Name:  "region, r",
			Usage: "AWS Region to use",
		},
		cli.StringFlag{
			Name:  "profile, p",
			Usage: "AWS config profile to use",
		},
		cli.StringFlag{
			Name:  "namespace, n",
			Usage: "Namespace to use as a prefix for stacks",
		},
		cli.BoolFlag{
			Name:  "silent, s",
			Usage: "silent mode, errors only",
		},
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "increase level of log verbosity",
		},
		cli.BoolFlag{
			Name:  "dryrun, d",
			Usage: "generate the cloudformation templates without upserting stacks",
		},
		cli.BoolFlag{
			Name:  "disable-iam, I",
			Usage: "disable the automatic creation of IAM resources",
		},
	}

	return app
}
