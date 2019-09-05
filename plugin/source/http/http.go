/*
	Package http implements the built-in HTTP download source for `go2chef`. It
	can download either single files or any archive format supported by github.com/mholt/archiver.
*/

package http

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/facebookincubator/go2chef/plugin/lib/certs"

	"github.com/facebookincubator/go2chef/util/temp"

	"github.com/facebookincubator/go2chef"
	"github.com/mholt/archiver"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this source plugin
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

	tlsConf, err := certs.TLS.GetTLSClientConf()
	if err != nil {
		return err
	}
	// Use the client from GlobalConfig so we can get any configured CAs
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConf,
		},
	}

	if ex, err := go2chef.PathExists(dlPath); err != nil {
		return err
	} else if !ex {
		s.logger.Debugf(1, "creating download directory %s", dlPath)
		if err := os.MkdirAll(dlPath, 0755); err != nil {
			return err
		}
	}
	s.logger.Debugf(1, "%s: 1", s.Name())

	req, err := http.NewRequest(s.Method, s.URL, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	s.logger.Debugf(1, "%s: HTTP %s %s => %d %s", s.Name(), s.Method, s.URL, resp.StatusCode, http.StatusText(resp.StatusCode))

	if !s.checkStatusCode(resp) {
		return fmt.Errorf("non-matching status code: %d", resp.StatusCode)
	}

	tmpfile, err := temp.File("", "go2chef-src-http-*")
	defer func() { _ = tmpfile.Close() }()
	if err != nil {
		return err
	}
	if _, err = io.Copy(tmpfile, resp.Body); err != nil {
		return err
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
		if err == nil {
			if fn, ok := params["filename"]; ok {
				outputFilename = fn
			}
		}
	}
	outputPath := filepath.Join(dlPath, outputFilename)

	if s.Archive {
		/*
		  ARCHIVE MODE: If the request is for an archive (using `{"archive": true}` in config) then
		  decompress that archive into the destination.
		*/
		_ = tmpfile.Close()
		s.logger.Debugf(1, "%s: archive mode enabled, extracting %s to %s", s.Name(), tmpfile.Name(), dlPath)
		extFilename := filepath.Join(filepath.Dir(tmpfile.Name()), outputFilename)
		if err := os.Rename(tmpfile.Name(), extFilename); err != nil {
			s.logger.Errorf("failed to relocate output")
			return err
		}

		if err := archiver.Unarchive(extFilename, dlPath); err != nil {
			return err
		}
	} else {
		/*
			FILE MODE: If the request isn't for an archive (default), then just close the temp file
			and move to the output path.
		*/
		s.logger.Debugf(1, "%s: direct download", s.Name())
		_ = tmpfile.Close()
		return os.Rename(tmpfile.Name(), outputPath)
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
