package environments

import(
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)

// NewListCommand returns a cli.Command to list environments
func NewListCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "list",
		Aliases: []string{"ls"},
		Usage: "list environments",
		Action: func(c *cli.Context) error {
			runList(config)
			return nil
		},
	}

	return cmd
}

func runList(config *common.Config) {
	fmt.Println("listing environments")
}
