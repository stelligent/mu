package environments

import(
	"fmt"
	"github.com/urfave/cli"
	"github.com/stelligent/mu/common"
)

func newUpsertCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "upsert",
		Aliases: []string{"up"},
		Usage: "create/update an environment",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			runUpsert(config, c.Args().First())
			return nil
		},
	}

	return cmd
}

func runUpsert(config *common.Config, environment string) {
	fmt.Printf("upserting environment: %s\n",environment)
}