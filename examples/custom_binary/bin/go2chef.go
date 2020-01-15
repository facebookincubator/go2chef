package main

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"github.com/facebookincubator/go2chef/cli"
	"os"
)

func main() {
	g2c := cli.NewGo2ChefCLI()
	exit := g2c.Run(os.Args)
	os.Exit(exit)
}
