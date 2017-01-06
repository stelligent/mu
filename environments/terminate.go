package environments

import(
	"fmt"
	"github.com/urfave/cli"
	"github.com/stelligent/mu/common"
)

// NewShowCommand returns a cli.Command to show environments
func NewTerminateCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "terminate",
		Aliases: []string{"term"},
		Usage: "terminate an environment",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			runTerminate(config, c.Args().First())
			return nil
		},
	}

	return cmd
}

func runTerminate(config *common.Config, environment string) {
	fmt.Printf("terminating environment: %s\n",environment)
}
