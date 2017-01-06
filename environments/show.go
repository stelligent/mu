package environments

import(
	"fmt"
	"github.com/urfave/cli"
	"github.com/stelligent/mu/common"
)

// NewShowCommand returns a cli.Command to show environments
func NewShowCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "show",
		Usage: "show environment details",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			runShow(config, c.Args().First())
			return nil
		},
	}

	return cmd
}

func runShow(config *common.Config, environment string) {
	fmt.Printf("showing environment: %s\n",environment)
}