package main

import (
	"github.com/stelligent/mu/cli"
	"github.com/stelligent/mu/common"
	"os"
)

var version string

func main() {
	common.SetVersion(version)
	cli.NewApp().Run(os.Args)
}
