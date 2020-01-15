package testutil

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

// ZipDirHandler provides a zip-and-ship HTTP handler
type ZipDirHandler struct {
	Root string
}

func (z *ZipDirHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	archiveName := path.Base(req.URL.Path) + ".zip"
	cmd := exec.Command("zip", "-r", "-", ".")
	cmd.Dir = filepath.Join(z.Root, req.URL.Path)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("error starting zip: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+archiveName+".zip")
	w.Header().Set("Content-Type", "application/zip")

	if err := cmd.Wait(); err != nil {
		log.Printf("error zipping files: %s", err)
		return
	}
	log.Printf("zipped request %s", req.URL.Path)
}
