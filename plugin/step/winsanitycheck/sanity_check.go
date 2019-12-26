package winsanitycheck

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"errors"
	"fmt"
	"log"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this plugin
const TypeName = "go2chef.sanitycheck"

// SanityCheck implements basic sanity checks, and provides an API
// for additional sanity-checks to be built as plugins.
type SanityCheck struct {
	SName   string   `mapstructure:"name"`
	Enabled []string `mapstructure:"enabled"`
}

func (s *SanityCheck) String() string {
	return "<" + TypeName + ":" + s.SName + ">"
}

// Name returns this step's name
func (s *SanityCheck) Name() string { return s.SName }

// Type returns "sanitycheck"
func (s *SanityCheck) Type() string { return TypeName }

// SetName sets this step's name
func (s *SanityCheck) SetName(n string) { s.SName = n }

// Download is an noop for sanity checking
func (s *SanityCheck) Download() error { return nil }

// Execute performs the sanity checks
func (s *SanityCheck) Execute() error {
	log.Printf("executing sanity checks")
	checks := make(map[string]CheckFn)
	for _, en := range s.Enabled {
		if sc, ok := sanityCheckRegistry[en]; ok {
			checks[en] = sc
		} else {
			return fmt.Errorf("sanity check %s does not exist", en)
		}
	}
	for n, c := range checks {
		log.Printf("running sanity check %s", n)
		fix, err := c(s)
		if err == ErrSanityCheckNeedsFix && fix != nil {
			if err := fix(s); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// Loader implements the go2chef.StepLoader interface required for plugins
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	sc := &SanityCheck{}
	if err := mapstructure.Decode(config, sc); err != nil {
		return nil, err
	}
	return sc, nil
}

var _ go2chef.Step = &SanityCheck{}
var _ go2chef.StepLoader = Loader

func init() {
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterStep(TypeName, Loader)
	}
	RegisterSanityCheck("superuser", EnsureSuperuser)
}

// CheckFn is the type for sanity checks
type CheckFn func(sc *SanityCheck) (FixFn, error)

// FixFn is the type for sanity check fixes
type FixFn func(sc *SanityCheck) error

// ErrSanityCheckNeedsFix is the error raised when a sanity check needs a fix run
var ErrSanityCheckNeedsFix = errors.New("sanity check needs fix")

var sanityCheckRegistry = make(map[string]CheckFn)

// RegisterSanityCheck registers an additional sanity check
func RegisterSanityCheck(name string, fn CheckFn) {
	if _, ok := sanityCheckRegistry[name]; ok {
		panic("sanity check " + name + " is already registered")
	}
	sanityCheckRegistry[name] = fn
}
