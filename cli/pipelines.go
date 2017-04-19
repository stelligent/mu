package cli

import (
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/workflows"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
	"time"
)

func newPipelinesCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "pipeline",
		Usage: "options for managing pipelines",
		Subcommands: []cli.Command{
			*newPipelinesListCommand(ctx),
			*newPipelinesUpsertCommand(ctx),
			*newPipelinesTerminateCommand(ctx),
			*newPipelinesLogsCommand(ctx),
		},
	}

	return cmd
}

func newPipelinesListCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "list pipelines",
		Action: func(c *cli.Context) error {
			workflow := workflows.NewPipelineLister(ctx, os.Stdout)
			return workflow()
		},
	}

	return cmd
}

func newPipelinesTerminateCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:      "terminate",
		Aliases:   []string{"term"},
		Usage:     "terminate pipeline",
		ArgsUsage: "[<service>]",
		Action: func(c *cli.Context) error {
			service := c.Args().First()
			workflow := workflows.NewPipelineTerminator(ctx, service)
			return workflow()
		},
	}

	return cmd
}

func newPipelinesUpsertCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:    "upsert",
		Aliases: []string{"up"},
		Usage:   "upsert pipeline",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "token, t",
				Usage: "GitHub token ",
			},
		},
		Action: func(c *cli.Context) error {
			token := c.String("token")
			workflow := workflows.NewPipelineUpserter(ctx, func(required bool) string {
				if required && token == "" {
					fmt.Print("  GitHub token: ")
					byteToken, err := terminal.ReadPassword(int(syscall.Stdin))
					if err == nil {
						token = strings.TrimSpace(string(byteToken))
						fmt.Println("")
					}
				}

				return token
			})
			return workflow()
		},
	}

	return cmd
}
func newPipelinesLogsCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "logs",
		Usage: "show pipeline logs",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: "service name to view logs for",
			},
			cli.BoolFlag{
				Name:  "follow, f",
				Usage: "follow logs for latest changes",
			},
			cli.DurationFlag{
				Name:  "search-duration, t",
				Usage: "duration to go into the past for searching (e.g. 5m for 5 minutes)",
				Value: 1 * time.Minute,
			},
		},
		ArgsUsage: "[<filter>...]",
		Action: func(c *cli.Context) error {
			serviceName := c.String("service")

			workflow := workflows.NewPipelineLogViewer(ctx, c.Duration("search-duration"), c.Bool("follow"), serviceName, os.Stdout, strings.Join(c.Args(), " "))
			return workflow()
		},
	}

	return cmd
}
