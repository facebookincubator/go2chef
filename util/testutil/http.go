package testutil

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ZipDirHandler provides a zip-and-ship HTTP handler
type ZipDirHandler struct {
	Root string
}

func (z *ZipDirHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		buf         bytes.Buffer
		path        = filepath.Join(z.Root, req.URL.Path)
		archiveName = func() string {
			_, leaf := filepath.Split(z.Root)
			return leaf + ".zip"
		}()
		zipArchive = zip.NewWriter(&buf)
	)

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		f, err := zipArchive.Create(info.Name())
		if err != nil {
			return err
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err := f.Write(b); err != nil {
			return err
		}

		return nil
	})

	if err := zipArchive.Close(); err != nil {
		log.Printf("error during zip creation: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+archiveName)
	w.Header().Set("Content-Type", "application/zip")
	w.Write(buf.Bytes())
	log.Printf("zipped folder %s for endpoint %s", path, req.URL.Path)
}
