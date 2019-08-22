/*
	Package http implements the built-in HTTP download source for `go2chef`. It
	can download either single files or any archive format supported by github.com/mholt/archiver.
*/

package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/facebookincubator/go2chef"
	"github.com/mholt/archiver"
	"github.com/mitchellh/mapstructure"
)

const TypeName = "go2chef.source.http"

// Source implements an HTTP source for resource downloads
type Source struct {
	logger           go2chef.Logger
	SourceName       string `mapstructure:"name"`
	Method           string `mapstructure:"http_method"`
	URL              string `mapstructure:"url"`
	ValidStatusCodes []int  `mapstructure:"valid_status_codes"`
	Archive          bool   `mapstructure:"archive"`
	OutputFilename   string `mapstructure:"output_filename"`
}

// String returns a string representation of this
func (s *Source) String() string {
	return "<"
}

// Name returns the name of this source
func (s *Source) Name() string {
	return s.SourceName
}

// Type returns "http"
func (s *Source) Type() string {
	return "http"
}

// SetName sets the name of this source
func (s *Source) SetName(n string) {
	s.SourceName = n
}

// DownloadToPath downloads a file over HTTP to a given path, handling
// archive extraction if the Source.Archive parameter is true.
func (s *Source) DownloadToPath(dlPath string) (err error) {
	// set up start/end events
	s.logger.WriteEvent(go2chef.NewEvent("HTTP_DOWNLOAD_STARTED", TypeName, s.URL))
	defer func() {
		event := "HTTP_DOWNLOAD_COMPLETE"
		if err != nil {
			event = "HTTP_DOWNLOAD_FAILURE"
		}
		s.logger.WriteEvent(go2chef.NewEvent(event, TypeName, s.URL))
	}()

	// Use the client from GlobalConfig so we can get any configured CAs
	c := go2chef.Global.GetHTTPClientWithCAs()

	if ex, err := go2chef.PathExists(dlPath); err != nil {
		return err
	} else if !ex {
		s.logger.D(3).Debugf("creating download directory %s", dlPath)
		if err := os.MkdirAll(dlPath, 0755); err != nil {
			return err
		}
	}

	req, err := http.NewRequest(s.Method, s.URL, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	s.logger.D(2).Debugf("%s: HTTP %s %s => %d %s", s.Name(), s.Method, s.URL, resp.StatusCode, http.StatusText(resp.StatusCode))

	if !s.checkStatusCode(resp) {
		return fmt.Errorf("non-matching status code: %d", resp.StatusCode)
	}

	/*
	  FILENAME DETERMINATION

	  Filenames are determined like so:
	  - Start with the basename of the request URL path
	  - Check if config["output_filename"] is set and use it if so
	  - If not, check if the Content-Disposition has a download filename set and use that if so
	*/
	outputFilename := path.Base(req.URL.Path)
	if s.OutputFilename != "" {
		outputFilename = dlPath
	} else {
		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err != nil {
			return err
		}
		if fn, ok := params["filename"]; ok {
			outputFilename = fn
		}
	}
	outputPath := filepath.Join(dlPath, outputFilename)

	if s.Archive {
		/*
		  ARCHIVE MODE

		  If the request is for an archive (using `{"archive": true}` in config) then
		  decompress that archive into the destination.
		*/
		s.logger.D(1).Debugf("%s: archive mode enabled", s.Name())
		tmpfile, err := ioutil.TempFile("", "go2chef-src-http-*-"+outputFilename)
		defer func() { _ = tmpfile.Close() }()
		if err != nil {
			return err
		}
		if _, err = io.Copy(tmpfile, resp.Body); err != nil {
			return err
		}
		s.logger.D(2).Debugf("%s: extracting %s to %s", s.Name(), tmpfile.Name(), dlPath)

		if err := archiver.Unarchive(tmpfile.Name(), dlPath); err != nil {
			return err
		}
	} else {
		/*
		  FILE MODE

		  If the request isn't for an archive (default), then just put it straight into
		  the destination. Step implementations can handle the file discovery logic
		  themselves.
		*/
		fh, err := os.OpenFile(outputPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(0644))
		if err != nil {
			return err
		}
		defer func() { _ = fh.Close() }()

		if _, err := io.Copy(fh, resp.Body); err != nil {
			return err
		}
	}
	return nil
}

// checkStatusCodes does the logic for checking if non-200 status codes
// were marked as okay in config.
func (s *Source) checkStatusCode(resp *http.Response) bool {
	if resp.StatusCode != 200 {
		for _, code := range s.ValidStatusCodes {
			if resp.StatusCode == code {
				return true
			}
		}
		return false
	}
	return true
}

// Loader implements SourceLoader for plugin registration
func Loader(config map[string]interface{}) (go2chef.Source, error) {
	s := &Source{
		go2chef.GetGlobalLogger(),
		"",
		"GET",
		"",
		make([]int, 0),
		false,
		"",
	}
	if err := mapstructure.Decode(config, s); err != nil {
		return nil, err
	}
	if s.SourceName == "" {
		s.SourceName = "http"
	}
	return s, nil
}

var _ go2chef.Source = &Source{}
var _ go2chef.SourceLoader = Loader

func init() {
	go2chef.RegisterSource(TypeName, Loader)
}
