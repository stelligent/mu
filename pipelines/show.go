package pipelines

import(
	"fmt"
	"github.com/urfave/cli"
	"github.com/stelligent/mu/common"
)

func newShowCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command {
		Name: "show",
		Usage: "show pipeline details",
		Flags: []cli.Flag {
			cli.StringFlag{
				Name: "service, s",
				Usage: "service to show",
			},
		},
		Action: func(c *cli.Context) error {
			runShow(config, c.String("service"))
			return nil
		},
	}

	return cmd
}

func runShow(config *common.Config, service string) {
	fmt.Printf("showing pipeline: %s\n",service)
}
