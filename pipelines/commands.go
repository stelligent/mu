package pipelines

import(
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

// NewCommand returns a cli.Command with all the pipeline subcommands
func NewCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "pipeline",
		Usage: "options for managing pipelines",
		Subcommands: []cli.Command{
			*newListCommand(config),
			*newShowCommand(config),
		},
	}

	return cmd
}

