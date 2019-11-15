package multi

import (
	"io/ioutil"
	"os"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
	"github.com/otiai10/copy"
)

// TypeName is the name of this source plugin
const TypeName = "go2chef.source.multi"

// Source implements a multi filesystem source plugin that
// copies files from a directory on multi filesystem to temp
// for use. Paths are relative to the current working directory
// of the go2chef process.
type Source struct {
	logger go2chef.Logger

	SourceName string `mapstructure:"name"`

	SourceSpecs []map[string]interface{} `mapstructure:"sources"`
	sources     []go2chef.Source         `mapstructure:","`
}

func (s *Source) String() string {
	return "<" + TypeName + ":" + s.SourceName + ">"
}

// Name returns the name of this source instance
func (s *Source) Name() string {
	return s.SourceName
}

// Type returns the type of this source
func (s *Source) Type() string {
	return TypeName
}

// SetName sets the name of this source instance
func (s *Source) SetName(name string) {
	s.SourceName = name
}

// DownloadToPath fetches multiple source defs in order
func (s *Source) DownloadToPath(dlPath string) error {
	if err := os.MkdirAll(dlPath, 0755); err != nil {
		return err
	}
	for i, src := range s.sources {
		thisDl, err := ioutil.TempDir("", "go2chef-multi-")
		if err != nil {
			return err
		}
		if err := src.DownloadToPath(thisDl); err != nil {
			s.logger.Errorf("failed to download source %d to %s", i, thisDl)
			return err
		}
		if err := copy.Copy(thisDl, dlPath); err != nil {
			s.logger.Errorf("failed to copy directory %s to %s", thisDl, dlPath)
			return err
		}
	}
	return nil
}

// Loader provides an instantiation function for this source
func Loader(config map[string]interface{}) (go2chef.Source, error) {
	s := &Source{
		logger:      go2chef.GetGlobalLogger(),
		SourceName:  "",
		SourceSpecs: []map[string]interface{}{},
	}
	if err := mapstructure.Decode(config, s); err != nil {
		return nil, err
	}
	if s.SourceName == "" {
		s.SourceName = TypeName
	}

	for _, spec := range s.SourceSpecs {
		stype, err := go2chef.GetType(spec)
		if err != nil {
			return nil, err
		}
		src, err := go2chef.GetSource(stype, spec)
		if err != nil {
			return nil, err
		}
		s.sources = append(s.sources, src)
	}

	return s, nil
}

var _ go2chef.Source = &Source{}
var _ go2chef.SourceLoader = Loader

func init() {
	go2chef.RegisterSource(TypeName, Loader)
}
