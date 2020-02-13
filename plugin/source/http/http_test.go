package http

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/facebookincubator/go2chef/util/testutil"
)

// Test behavior with regular files
func TestSource_DownloadToPath(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %s", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment; filename=test.txt")
		_, _ = fmt.Fprint(w, "hello")
	}))
	defer ts.Close()

	s, err := Loader(map[string]interface{}{
		"url": ts.URL,
	})
	if err != nil {
		t.Fatalf("failed to initialize source: %s", err)
	}
	if err := s.DownloadToPath(dir); err != nil {
		t.Errorf("failed to download from %s to path %s: %s", ts.URL, dir, err)
	}
	dlpath := filepath.Join(dir, "test.txt")
	if data, err := ioutil.ReadFile(dlpath); err != nil {
		t.Errorf("failed to read downloaded file from %s: %s", dlpath, err)
	} else {
		if string(data) != "hello" {
			t.Errorf("did not get expected content `hello`: %#v", data)
		}
	}
}

// Test behavior with archive=true
func TestSource_DownloadToPathArchive(t *testing.T) {
	testfileName := "test.txt"
	content := "hello world!"

	wdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %s", err)
	}

	testfileDir := filepath.Join(wdir, "test")

	if err := os.MkdirAll(testfileDir, 0755); err != nil {
		t.Fatalf("failed to create test archive directory: %s", err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(testfileDir, testfileName),
		[]byte(content),
		0644,
	); err != nil {
		t.Fatalf("failed to write temporary directory file")
	}

	ts := httptest.NewServer(&testutil.ZipDirHandler{Root: wdir})
	defer ts.Close()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("failed to create temporary directory: %s", err)
	}

	s, err := Loader(map[string]interface{}{
		"url":     ts.URL + "/test",
		"archive": true,
	})
	if err != nil {
		t.Fatalf("failed to initialize source: %s", err)
	}
	if err := s.DownloadToPath(dir); err != nil {
		t.Errorf("failed to download from %s to path %s: %s", ts.URL, dir, err)
	}
	dlpath := filepath.Join(dir, "test", testfileName)
	if data, err := ioutil.ReadFile(dlpath); err != nil {
		t.Errorf("failed to read downloaded file from %s: %s", dlpath, err)
	} else {
		if string(data) != content {
			t.Errorf("did not get expected content `%s`: %#v => `%s`", content, data, string(data))
		}
	}
}
