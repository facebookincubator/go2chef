package go2chef

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"github.com/spf13/pflag"
	"testing"
)

type DummyConfigSource struct{}

func (d *DummyConfigSource) InitFlags(set *pflag.FlagSet) {}
func (d *DummyConfigSource) ReadConfig() (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

var _ ConfigSource = &DummyConfigSource{}

func TestConfigSourceRegistry(t *testing.T) {
	RegisterConfigSource("dupe", &DummyConfigSource{})
	if !doesFunctionPanic(func() {
		RegisterConfigSource("dupe", nil)
	}) {
		t.Fatalf("RegisterConfigSource does not panic on duplicate")
	}

	if cs := GetConfigSource("dupe"); cs == nil {
		t.Errorf("failed to get config source `dupe` despite it being registered")
	}
}
