package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/facebookincubator/go2chef/cli"
)

var banner = `
          ___    _         __
 __ _ ___|_  )__| |_  ___ / _|
/ _` + "`" + ` / _ \/ // _| ' \/ -_)  _|
\__, \___/___\__|_||_\___|_|
|___/
`

func main() {
	for _, line := range strings.Split(banner, "\n") {
		_, _ = fmt.Fprintln(os.Stderr, line)
	}
	g2c := cli.NewGo2ChefCLI()
	exit := g2c.Run(os.Args)
	os.Exit(exit)
}
