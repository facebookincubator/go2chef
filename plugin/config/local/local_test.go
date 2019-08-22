package local

import (
	"io/ioutil"
	"testing"
)

func TestConfigSource(t *testing.T) {
	cs := &ConfigSource{}

	tf := func(t *testing.T, content string) string {
		tf, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatalf("error creating tempfile: %s", err)
		}
		defer tf.Close()
		if _, err := tf.WriteString(content); err != nil {
			t.Fatalf("failed to write tempfile: %s", err)
		}
		return tf.Name()
	}(t, `{"key":"value"}`)

	cs.Path = tf
	if cr, err := cs.ReadConfig(); err != nil {
		t.Fatalf("failed to read config: %s", err)
	} else {
		if v, ok := cr["key"]; ok {
			if v != "value" {
				t.Errorf("config[key] != value")
			}
		} else {
			t.Errorf("config[key] does not exist")
		}
	}
}
