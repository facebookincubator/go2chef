package local

import (
	"encoding/json"
	"github.com/facebookincubator/go2chef"
	"github.com/spf13/pflag"
	"io/ioutil"
)

const TypeName = "go2chef.config_source.local"

type ConfigSource struct {
	Path string
}

func (c *ConfigSource) InitFlags(set *pflag.FlagSet) {
	set.StringVar(&c.Path, "local-config", "", "local configuration path")
}

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
