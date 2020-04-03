package depnotify

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"os"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this step plugin
const TypeName = "go2chef.step.depnotify"

// Step implements a depnotify execution step plugin
type Step struct {
	SName   string `mapstructure:"name"`
	Status  bool
	Message string
	logger  go2chef.Logger
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

// Download does nothing for this step since there's no
// downloading to be done when running any ol' command.
func (s *Step) Download() error {
	return nil
}

// Execute appends the status to the depnotify log file.
func (s *Step) Execute() error {
	prefix := "Command: "
	if s.Status {
		prefix = "Status: "
	}
	message := prefix + s.Message + "\n"
	f, err := os.OpenFile("/private/var/tmp/depnotify.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(message); err != nil {
		return err
	}
	return nil
}

// Loader provides an instantiation function for this step plugin
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	c := &Step{
		logger: go2chef.GetGlobalLogger(),
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
