package file

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this step plugin
const TypeName = "go2chef.step.file"

// Step implements a command execution step plugin
type Step struct {
	SName        string `mapstructure:"name"`
	source       go2chef.Source
	logger       go2chef.Logger
	DownloadPath string `mapstructure:"path"`
}

func (s *Step) String() string {
	return "<" + TypeName + ":" + s.SName + ">"
}

// SetName sets the name of this step instance
func (s *Step) SetName(name string) {
	s.SName = name
}

// Name returns the name of this step instance
func (s *Step) Name() string {
	return s.SName
}

// Type returns the type of this step instance
func (s *Step) Type() string {
	return TypeName
}

// Download is the whole point since we're really just
// placing a file.
func (s *Step) Download() error {
	if s.source == nil {
		return nil
	}
	s.logger.Debugf(1, "%s: downloading source to path: %s", s.Name(), s.DownloadPath)

	if err := s.source.DownloadToPath(s.DownloadPath); err != nil {
		return err
	}

	s.logger.Debugf(1, "%s: downloaded source to %s", s.Name(), s.DownloadPath)
	return nil
}

// Execute does nothing right now
func (s *Step) Execute() error {
	// potentially use the execute step to set mode?
	return nil
}

// Loader provides an instantiation function for this step plugin
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	source, err := go2chef.GetSourceFromStepConfig(config)
	if err != nil {
		return nil, err
	}
	c := &Step{
		logger: go2chef.GetGlobalLogger(),
		source: source,
	}
	if err := mapstructure.Decode(config, c); err != nil {
		return nil, err
	}

	return c, nil
}

var _ go2chef.Step = &Step{}
var _ go2chef.StepLoader = Loader

func init() {
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterStep(TypeName, Loader)
	}
}
