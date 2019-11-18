package main

import (
	_ "github.com/facebookincubator/go2chef/plugin/config/http"
	_ "github.com/facebookincubator/go2chef/plugin/config/local"
	_ "github.com/facebookincubator/go2chef/plugin/logger/stdlib"
	_ "github.com/facebookincubator/go2chef/plugin/source/http"
	_ "github.com/facebookincubator/go2chef/plugin/source/local"
	_ "github.com/facebookincubator/go2chef/plugin/source/multi"
	_ "github.com/facebookincubator/go2chef/plugin/step/bundle"
	_ "github.com/facebookincubator/go2chef/plugin/step/command"
	_ "github.com/facebookincubator/go2chef/plugin/step/group"
	_ "github.com/facebookincubator/go2chef/plugin/step/install/darwin/pkg"
	_ "github.com/facebookincubator/go2chef/plugin/step/install/linux/apt"
	_ "github.com/facebookincubator/go2chef/plugin/step/install/linux/dnf"
	_ "github.com/facebookincubator/go2chef/plugin/step/install/windows/msi"
	_ "github.com/facebookincubator/go2chef/plugin/step/sanitycheck"
)
