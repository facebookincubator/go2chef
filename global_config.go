package go2chef

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

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
