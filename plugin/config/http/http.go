package http

import (
	"encoding/json"
	"net/http"

	"github.com/facebookincubator/go2chef"
	"github.com/spf13/pflag"
)

// TypeName is the name of this configuration source
const TypeName = "go2chef.config_source.http"

// ConfigSource loads configuration data from JSON files from an http source
type ConfigSource struct {
	URL string
}

// InitFlags sets the command-line flags for http configuration sources
func (c *ConfigSource) InitFlags(set *pflag.FlagSet) {
	set.StringVar(&c.URL, "http-config", "", "http configuration path")
}

// ReadConfig loads the configuration file from http
func (c *ConfigSource) ReadConfig() (map[string]interface{}, error) {
	r, err := http.Get(c.URL)
	if err != nil {
		return nil, err
	}
	output := make(map[string]interface{})
	if err := json.NewDecoder(r.Body).Decode(&output); err != nil {
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
