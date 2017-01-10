package main

import (
    "os"
    "github.com/stelligent/mu/cli"
)

var version string

func main() {
    cli.NewApp(version).Run(os.Args)
}
