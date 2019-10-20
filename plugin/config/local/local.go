package local

import (
	"encoding/json"
	"io/ioutil"

	"github.com/facebookincubator/go2chef"
	"github.com/spf13/pflag"
)

// TypeName is the name of this configuration source
const TypeName = "go2chef.config_source.local"

// ConfigSource loads configuration data from JSON files on the local filesystem
type ConfigSource struct {
	Path string
}

// InitFlags sets the command-line flags for local configuration sources
func (c *ConfigSource) InitFlags(set *pflag.FlagSet) {
	set.StringVar(&c.Path, "local-config", "config.json", "local configuration path")
}

// ReadConfig loads the configuration file from disk
func (c *ConfigSource) ReadConfig() (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(c.Path)
	if err != nil {
		return nil, err
	}
	output := make(map[string]interface{})
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}
	return output, nil
}

var _ go2chef.ConfigSource = &ConfigSource{}

func init() {
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterConfigSource(TypeName, &ConfigSource{})
	}
}
