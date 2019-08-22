package main

import (
	"os"

	"github.com/facebookincubator/go2chef/cli"

	_ "noop"

	_ "github.com/facebookincubator/go2chef/plugin/config/local"
)

func main() {
	c := cli.NewGo2ChefCLI()
	os.Exit(c.Run(os.Args))
}
