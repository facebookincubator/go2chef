// Package embed is a configuration source that can be fully compiled-in
package embed

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"github.com/facebookincubator/go2chef"
	"github.com/spf13/pflag"
)

// TypeName is the name of this config source
const TypeName = "go2chef.config_source.embed"

// ConfigSource is the embedded configuration source implementation
type ConfigSource struct{}

// InitFlags initializes flags for this config source (none)
func (c *ConfigSource) InitFlags(set *pflag.FlagSet) {}

// EmbeddedConfig exposes the means for storing the embedded configuration.
// If you want to embed configuration in some other format you can set this
// variable in an init() function in your own package to parse/store it.
var EmbeddedConfig = make(map[string]interface{})

// ReadConfig reads the configuration source
func (c *ConfigSource) ReadConfig() (map[string]interface{}, error) {
	return EmbeddedConfig, nil
}

var _ go2chef.ConfigSource = &ConfigSource{}

func init() {
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterConfigSource(TypeName, &ConfigSource{})
	}
}
