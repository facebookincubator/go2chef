package temp

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
)

var tmps = make(map[string]string)

// Dir creates a temporary directory registered for cleanup
func Dir(dir, prefix string) (name string, err error) {
	name, err = ioutil.TempDir(dir, prefix)
	if err == nil {
		_, fn, ln, _ := runtime.Caller(1)
		tmps[name] = fmt.Sprintf("%s:%d", fn, ln)
	}
	return
}

// File creates a temporary file registered for cleanup
func File(dir string, prefix string) (f *os.File, err error) {
	f, err = ioutil.TempFile(dir, prefix)
	if err == nil {
		_, fn, ln, _ := runtime.Caller(1)
		tmps[f.Name()] = fmt.Sprintf("%s:%d", fn, ln)
	}
	return
}

// Cleanup performs the cleanup of paths unless preserve is true, in which
// case it just prints the paths without deleting them (for debugging).
func Cleanup(preserve bool) {
	var errs []error
	for tmp, caller := range tmps {
		if preserve {
			log.Printf("preserving temp path %s", tmp)
		} else if err := os.RemoveAll(tmp); err != nil {
			log.Printf("util.temp.Cleanup(%s) failed: %s (created by %s)", tmp, err, caller)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		strs := make([]string, 0, len(errs))
		for _, e := range errs {
			strs = append(strs, e.Error())
		}
		log.Printf("temp dirs cleanup completed with errors: %s", strings.Join(strs, "\n"))
	}
	log.Printf("temp dirs cleanup completed")
}
