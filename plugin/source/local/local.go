package local

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"os"

	"github.com/mholt/archiver"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
	"github.com/otiai10/copy"
)

// TypeName is the name of this source plugin
const TypeName = "go2chef.source.local"

// Source implements a local filesystem source plugin that
// copies files from a directory on local filesystem to temp
// for use. Paths are relative to the current working directory
// of the go2chef process.
type Source struct {
	logger go2chef.Logger

	SourceName string `mapstructure:"name"`
	Path       string `mapstructure:"path"`
	Archive    bool   `mapstructure:"archive"`
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

// DownloadToPath performs the actual copy of files to the working directory.
// We copy rather than just setting downloadPath to avoid side effects from
// steps affecting the original source location.
func (s *Source) DownloadToPath(dlPath string) error {

	if err := os.MkdirAll(dlPath, 0755); err != nil {
		return err
	}
	s.logger.Debugf(0, "copy directory %s is ready", dlPath)

	if !s.Archive {
		if err := copy.Copy(s.Path, dlPath); err != nil {
			s.logger.Errorf("failed to copy directory %s to %s", s.Path, dlPath)
			return err
		}
		s.logger.Debugf(0, "copied directory %s to %s", s.Path, dlPath)
	} else {
		if err := archiver.Unarchive(s.Path, dlPath); err != nil {
			s.logger.Errorf("failed to unarchive %s to dir %s", s.Path, dlPath)
			return err
		}
	}

	return nil
}

// Loader provides an instantiation function for this source
func Loader(config map[string]interface{}) (go2chef.Source, error) {
	s := &Source{
		logger:     go2chef.GetGlobalLogger(),
		SourceName: "",
	}
	if err := mapstructure.Decode(config, s); err != nil {
		return nil, err
	}
	if s.SourceName == "" {
		s.SourceName = TypeName
	}
	return s, nil
}

var _ go2chef.Source = &Source{}
var _ go2chef.SourceLoader = Loader

func init() {
	go2chef.RegisterSource(TypeName, Loader)
}
