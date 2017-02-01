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
)

func newPipelinesCommand(ctx *common.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "pipeline",
		Usage: "options for managing pipelines",
		Subcommands: []cli.Command{
			*newPipelinesListCommand(ctx),
			*newPipelinesUpsertCommand(ctx),
			*newPipelinesTerminateCommand(ctx),
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
		Name:    "terminate",
		Aliases: []string{"term"},
		Usage:   "terminate pipeline",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "service, s",
				Usage: "service to terminate pipeline",
			},
		},
		Action: func(c *cli.Context) error {
			service := c.String("service")
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
					}
				}

				return token
			})
			return workflow()
		},
	}

	return cmd
}
