/*
  GLOBAL CONFIGURATION

  This file implements go2chef's global configuration subsystem. This is not implemented
  with a plugin model as plugins which need functionality not yet provided here can still
  implement it within their own plugin config.
*/

package go2chef

import (
	"github.com/facebookincubator/go2chef/util/plugconf"
	"github.com/mitchellh/mapstructure"
)

var GlobalConfiguration = plugconf.NewPlugConf()

func LoadGlobalConfiguration(config map[string]interface{}) error {
	gc, ok := config["global"]
	if !ok {
		return nil
	}
	gcmap := make(map[string]interface{})
	if err := mapstructure.Decode(gc, &gcmap); err != nil {
		return err
	}
	return GlobalConfiguration.Process(gcmap)
}
