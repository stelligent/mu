package pipelines

import(
	"fmt"
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
)


func newListCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "list",
		Aliases: []string{"ls"},
		Usage: "list pipelines",
		Action: func(c *cli.Context) error {
			runList(config)
			return nil
		},
	}

	return cmd
}

func runList(config *common.Config) {
	fmt.Println("listing pipelines")
}
