package main

import (
	"github.com/stelligent/mu/cli"
	"os"
)

var version string

func main() {
	cli.NewApp(version).Run(os.Args)
}
