/*
Package embed implements a configuration source that can be fully compiled-in
at build time.
*/

package embed

import (
	"github.com/facebookincubator/go2chef"
	"github.com/spf13/pflag"
)

const TypeName = "go2chef.config_source.embed"

type ConfigSource struct{}

func (c *ConfigSource) InitFlags(set *pflag.FlagSet) {}

var EmbeddedConfig = make(map[string]interface{})

func (c *ConfigSource) ReadConfig() (map[string]interface{}, error) {
	return EmbeddedConfig, nil
}

var _ go2chef.ConfigSource = &ConfigSource{}

func init() {
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterConfigSource(TypeName, &ConfigSource{})
	}
}
