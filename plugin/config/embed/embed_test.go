package embed

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import "testing"

func TestConfigSource(t *testing.T) {
	cs := &ConfigSource{}
	input := map[string]int{
		"a": 1,
		"b": 2,
	}
	EmbeddedConfig = map[string]interface{}{
		"a": 1,
		"b": 2,
	}
	read, err := cs.ReadConfig()
	if err != nil {
		t.Fatalf("somehow got an error calling ReadConfig(): %s", err)
	}
	for k, v := range read {
		switch v.(type) {
		case int:
			if input[k] != v {
				t.Errorf("mismatched output from ReadConfig(): %d != %d for %s", v, input[k], k)
			}
		}
	}
}
