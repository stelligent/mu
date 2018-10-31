package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/provider/aws"
	"github.com/urfave/cli"
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
		*newValidateCommand(context),
		*newEnvironmentsCommand(context),
		*newServicesCommand(context),
		*newDatabasesCommand(context),
		*newPipelinesCommand(context),
		*newCatalogCommand(context),
		*newPurgeCommand(context),
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
		log.Debugf("version:%v", common.GetVersion())

		// TODO: support initializing context from other cloud providers?
		log.Debugf("dryrun:%v path:%v", c.Bool("dryrun"), c.String("dryrun-output"))
		dryrunPath := ""
		if c.Bool("dryrun") {
			dryrunPath = c.String("dryrun-output")
		}
		err = aws.InitializeContext(context, c.String("profile"), c.String("assume-role"), c.String("region"), dryrunPath, c.Bool("skip-version-check"), c.String("proxy"), c.Bool("allow-data-loss"))
		if err != nil {
			return err
		}

		err = context.InitializeConfigFromFile(c.String("config"))
		if c.Args().First() != "init" {
			if err != nil {
				log.Warningf("Unable to load mu config: %v", err)
			}
			if err = context.Config.Validate(); err != nil {
				log.Errorf("Invalid Config: %v", err)
				return nil
			}
		}
		context.Config.DryRun = c.Bool("dryrun")

		// Allow overriding the `DisableIAM` in config via `--disable-iam` or `-I`
		if c.Bool("disable-iam") {
			context.Config.DisableIAM = true
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

		// initialize extensions
		return context.InitializeExtensions()
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
			Name:  "assume-role, a",
			Usage: "ARN of IAM role to assume",
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
		cli.StringFlag{
			Name:  "dryrun-output, O",
			Usage: "output directory for dryrun",
			Value: fmt.Sprintf("%s/mu-dryrun", os.TempDir()),
		},
		cli.BoolFlag{
			Name:  "disable-iam, I",
			Usage: "disable the automatic creation of IAM resources",
		},
		cli.BoolFlag{
			Name:  "skip-version-check, F",
			Usage: "disable the checking of stack major numbers before updating",
		},
		cli.StringFlag{
			Name:  "proxy, P",
			Usage: "Proxy to route AWS requests through",
		},
		cli.BoolFlag{
			Name:  "allow-data-loss",
			Usage: "temporarily allow delete or replace on RDS or KMS resources",
		},
	}

	return app
}
