package environments

import(
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

// NewCommand returns a cli.Command with all the environment subcommands
func NewCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "environment",
		Aliases: []string{"env"},
		Usage: "options for managing environments",
		Subcommands: []cli.Command{
			*newListCommand(config),
			*newShowCommand(config),
			*newUpsertCommand(config),
			*newTerminateCommand(config),
		},
	}

	return cmd
}
