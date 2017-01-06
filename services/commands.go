package services

import(
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

// NewCommand returns a cli.Command with all the service subcommands
func NewCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "service",
		Aliases: []string{"svc"},
		Usage: "options for managing services",
		Subcommands: []cli.Command{
			*newShowCommand(config),
			*newDeployCommand(config),
			*newSetenvCommand(config),
			*newUndeployCommand(config),
		},
	}

	return cmd
}






