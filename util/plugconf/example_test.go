package plugconf

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"testing"

	"github.com/mitchellh/mapstructure"
)

func TestPlugConfExampleSimple(t *testing.T) {
	/*
		Program-side code: create a PlugConf in your program, probably as an
		exported variable so plugins can register themselves from init()
	*/
	pc := testpc()

	/*
		Plugin-side code: create an output variable for your final config data
		and store the rest of the data.
	*/
	var output int
	proc := func(f string, data interface{}) error {
		if err := mapstructure.Decode(data, &output); err != nil {
			return err
		}
		return nil
	}
	/*
		More plugin-side code: register the field name you want the processor
		function to handle.
	*/
	if err := pc.Register("a", proc); err != nil {
		t.Fatalf("failed to register proc as field `proc`")
	}

	/*
		Program-side code: process actual configuration data
	*/
	ncd := newConfigData()
	if err := pc.Process(ncd); err != nil {
		t.Fatalf("error processing")
	}

	/*
		If everything works right, the processor function will set up the config
		state for whatever component registered it.
	*/
	exp := ncd["a"].(int)
	if output != exp {
		t.Errorf("output should be 1 after process, got %d", exp)
	}
}
