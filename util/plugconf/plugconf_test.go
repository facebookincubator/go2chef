package plugconf

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"log"
	"testing"
)

func testpc() *PlugConf {
	return NewPlugConf()
}

func newConfigData() map[string]interface{} {
	return map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": "car",
		"d": "dog",
		"e": 3e5,
		"f": false,
	}
}

func TestNewPlugConf(t *testing.T) {
	pc := NewPlugConf()
	if pc == nil {
		t.Errorf("NewPlugConf returned nil")
	}
}

func TestPlugConf_Register(t *testing.T) {
	pc := testpc()
	proc1 := func(f string, o interface{}) error {
		log.Printf("proc1 f: %s o: %#v", f, o)
		return nil
	}
	if err := pc.Register("proc1", proc1); err != nil {
		t.Fatalf("failed to register proc1: %s", err)
	}

	proc2 := func(f string, o interface{}) error {
		log.Printf("proc2 f: %s o: %#v", f, o)
		return nil
	}
	if err := pc.Register("proc1", proc2); err == nil {
		t.Fatalf("should have failed to register fn proc2 as field `proc1`")
	}

	if err := pc.Register("proc2", proc2); err != nil {
		t.Fatalf("failed to register proc2: %s", err)
	}

	var panic bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				panic = true
				t.Logf("panic caught successfully from MustRegister: %#v", r)
			}
		}()
		pc.MustRegister("proc2", proc2)
	}()
	if !panic {
		t.Fatalf("MustRegister(proc2) should panic")
	}
}

func TestPlugConf_Process(t *testing.T) {
	pc := testpc()
	proc1 := func(f string, o interface{}) error {
		log.Printf("proc1 f: %s o: %#v", f, o)
		return nil
	}
	if err := pc.Register("proc1", proc1); err != nil {
		t.Fatalf("failed to register proc1: %s", err)
	}
	if err := pc.Process(newConfigData()); err != nil {
		t.Fatalf("failed to process pc: %s", err)
	}
}
