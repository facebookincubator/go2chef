package testutil

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/mholt/archiver/v3"
)

// ZipDirHandler provides a zip-and-ship HTTP handler
type ZipDirHandler struct {
	Root string
}

func (z *ZipDirHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	archiveName := path.Base(req.URL.Path) + ".zip"
	zip := archiver.NewZip()

	defer func() {
		zip.Close()
		if err := os.Remove(archiveName); err != nil {
			log.Println(err, "removing", req.URL.Path)
			return
		}

		log.Printf("zipped request %s", req.URL.Path)
	}()

	zip.Archive([]string{filepath.Join(z.Root, req.URL.Path)}, archiveName)
	b, err := ioutil.ReadFile(archiveName)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+archiveName+".zip")
	w.Header().Set("Content-Type", "application/zip")
	w.Write(b)
}
