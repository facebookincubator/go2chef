package http

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigSource(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment; filename=test.txt")
		_, _ = fmt.Fprint(w, `{"key":"value"}`)
	}))
	defer ts.Close()

	cs := &ConfigSource{URL: ts.URL}

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
