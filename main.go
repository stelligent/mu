package main

import (
	"os"

	"github.com/stelligent/mu/cli"
	"github.com/stelligent/mu/common"
)

var version string

func main() {
	common.SetVersion(version)
	cli.NewApp().Run(os.Args)
}
