package local

import (
	"os"

	"github.com/mholt/archiver"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
	"github.com/otiai10/copy"
)

const TypeName = "go2chef.source.local"

type Source struct {
	logger go2chef.Logger

	SourceName string `mapstructure:"name"`
	Path       string `mapstructure:"path"`
	Archive    bool   `mapstructure:"archive"`
}

func (s *Source) String() string {
	return "<" + TypeName + ":" + s.SourceName + ">"
}

func (s *Source) Name() string {
	return s.SourceName
}

func (s *Source) Type() string {
	return TypeName
}

func (s *Source) SetName(name string) {
	s.SourceName = name
}

func (s *Source) DownloadToPath(dlPath string) error {

	if err := os.MkdirAll(dlPath, 0755); err != nil {
		return err
	}
	s.logger.Debugf("copy directory %s is ready", dlPath)

	if !s.Archive {
		if err := copy.Copy(s.Path, dlPath); err != nil {
			s.logger.Errorf("failed to copy directory %s to %s", s.Path, dlPath)
			return err
		}
		s.logger.Debugf("copied directory %s to %s", s.Path, dlPath)
	} else {
		if err := archiver.Unarchive(s.Path, dlPath); err != nil {
			s.logger.Errorf("failed to unarchive %s to dir %s", s.Path, dlPath)
			return err
		}
	}

	return nil
}

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
