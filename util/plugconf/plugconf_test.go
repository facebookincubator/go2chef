package plugconf

import "testing"

func TestNewPlugConf(t *testing.T) {
	pc := NewPlugConf(map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": "car",
		"d": "dog",
		"e": 3e5,
		"f": false,
	})
	if pc == nil {
		t.Errorf("NewPlugConf returned nil")
	}
}

func TestPlugConf_Register(t *testing.T) {
	pc := NewPlugConf(map[string]interface{}{
		"a": 1,
		"b": "2",
		"c": false,
	})
	st1 := struct{
		A int `plugconf:"a"`
	}{}
	if err := pc.Register(&st1); err != nil {
		t.Fatalf("failed to register st1: %s", err)
	}
	st2 := struct{
		B string `plugconf:"b"`
	}{}
	if err := pc.Register(&st2); err != nil {
		t.Fatalf("failed to register st2: %s", err)
	}
	st3 := struct{
		C bool `plugconf:"c"`
	}{}
	if err := pc.Register(&st3); err != nil {
		t.Fatalf("failed to register st3: %s", err)
	}
	st4 := struct{
		C bool `plugconf:"c"`
	}{}
	if err := pc.Register(&st4); err == nil {
		t.Fatalf("did not fail to register st3 but should have")
	}

}

func TestPlugConf_process(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		pc := NewPlugConf(map[string]interface{}{
			"a": 1,
			"b": "2",
			"c": false,
		})
		st := struct{
			A int `plugconf:"a"`
			B string `plugconf:"b"`
			C bool `plugconf:"c"`
		}{}
		if err := pc.process(&st); err != nil {
			t.Errorf("failed to process: %s", err)
		}
		if len(pc.remaining) > 0 {
			t.Errorf("expected all fields to be consumed")
		}
	})
	t.Run("two fields", func(t *testing.T) {
		pc := NewPlugConf(map[string]interface{}{
			"a": 1,
			"b": "2",
			"c": false,
		})
		st := struct{
			A int `plugconf:"a"`
			B string `plugconf:"b"`
		}{}
		if err := pc.process(&st); err != nil {
			t.Errorf("failed to process: %s", err)
		}
		if len(pc.remaining) != 1 {
			t.Errorf("expected all but 1 fields to be consumed")
		}
	})
	t.Run("one field", func(t *testing.T) {
		pc := NewPlugConf(map[string]interface{}{
			"a": 1,
			"b": "2",
			"c": false,
		})
		st := struct{
			A int `plugconf:"a"`
		}{}
		if err := pc.process(&st); err != nil {
			t.Errorf("failed to process: %s", err)
		}
		if len(pc.remaining) != 2 {
			t.Errorf("expected one field to be consumed")
		}
	})
	t.Run("no fields", func(t *testing.T) {
		pc := NewPlugConf(map[string]interface{}{
			"a": 1,
			"b": "2",
			"c": false,
		})
		st := struct{
		}{}
		if err := pc.process(&st); err != nil {
			t.Errorf("failed to process: %s", err)
		}
		if len(pc.remaining) != 3 {
			t.Errorf("expected no fields to be consumed")
		}
	})
	t.Run("no fields", func(t *testing.T) {
		pc := NewPlugConf(map[string]interface{}{
			"a": 1,
			"b": "2",
			"c": false,
		})
		st1 := struct{
			A int `plugconf:"a"`
			B string `plugconf:"b"`
		}{}
		st2 := struct{
			C bool `plugconf:"c"`
		}{}
		if err := pc.process(&st1); err != nil {
			t.Errorf("failed to process: %s", err)
		}
		if len(pc.remaining) != 1 {
			t.Errorf("expected two fields to be consumed")
		}
		if err := pc.process(&st2); err != nil {
			t.Errorf("failed to process second: %s", err)
		}
		if len(pc.remaining) > 0 {
			t.Errorf("expected all fields to be consumed")
		}
		if st1.A != 1 || st1.B != "2" || st2.C != false {
			t.Errorf("mismatched values in st1/st2: %#v %#v", st1, st2)
		}
	})
}

func TestPlugConf_Process(t *testing.T) {
	pc := NewPlugConf(map[string]interface{}{
		"a": 1,
		"b": "2",
		"c": false,
	})
	st1 := struct{
		A int `plugconf:"a"`
	}{}
	if err := pc.Register(&st1); err != nil {
		t.Fatalf("failed to register st1: %s", err)
	}
	st2 := struct{
		B string `plugconf:"b"`
	}{}
	if err := pc.Register(&st2); err != nil {
		t.Fatalf("failed to register st2: %s", err)
	}
	st3 := struct{
		C bool `plugconf:"c"`
	}{}
	if err := pc.Register(&st3); err != nil {
		t.Fatalf("failed to register st3: %s", err)
	}
	if err := pc.Process(); err != nil {
		t.Fatalf("failed to process pluggable config: %s", err)
	}
	if st1.A != 1 {
		t.Errorf("mismatched valud for st1.A: %d != 1", st1.A)
	}
	if st2.B != "2" {
		t.Errorf("mismatched valud for st2.B: %s != 1", st2.B)
	}
	if st3.C != false {
		t.Errorf("mismatched valud for st3.C: %t != false", st3.C)
	}
}