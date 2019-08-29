package msi

import (
	"context"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/facebookincubator/go2chef/util/temp"

	"github.com/facebookincubator/go2chef/util"

	"github.com/mitchellh/mapstructure"

	"github.com/facebookincubator/go2chef"
)

// TypeName is the name of this plugin
const TypeName = "go2chef.step.install.windows.msi"

var (
	// DefaultPackageName sets the default package name for MSI matchign
	DefaultPackageName = "chef"
)

// Step implements Chef installation via Windows MSI
type Step struct {
	StepName     string `mapstructure:"name"`
	MSIMatch     string `mapstructure:"msi_match"`
	ProgramMatch string `mapstructure:"program_match"`

	MSIEXECTimeoutSeconds int `mapstructure:"msiexec_timeout_seconds"`

	logger       go2chef.Logger
	source       go2chef.Source
	downloadPath string
}

func (s *Step) String() string { return "<" + TypeName + ":" + s.StepName + ">" }

// SetName sets the name of this step instance
func (s *Step) SetName(str string) { s.StepName = str }

// Name gets the name of this step instance
func (s *Step) Name() string { return s.StepName }

// Type returns the type of this step instance
func (s *Step) Type() string { return TypeName }

// Download fetches resources required for this step's execution
func (s *Step) Download() error {
	if s.source == nil {
		return nil
	}

	tmpdir, err := temp.Dir("", "go2chef-install")
	if err != nil {
		return err
	}

	if err := s.source.DownloadToPath(tmpdir); err != nil {
		return err
	}
	s.downloadPath = tmpdir

	return nil
}

// Execute performs the installation
func (s *Step) Execute() error {
	msi, err := s.findMSI()
	if err != nil {
		return err
	}

	instCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.MSIEXECTimeoutSeconds)*time.Second)
	defer cancel()

	// create a logfile for MSIEXEC
	logfile, err := temp.File("", "")
	if err != nil {
		return err
	}
	_ = logfile.Close()

	cmd := exec.CommandContext(instCtx, "msiexec", "/qn", "/i", filepath.Join(s.downloadPath, msi), "/L*V", logfile.Name())

	if err := cmd.Run(); err != nil {
		// preserve exit error
		xerr := err
		if exit, ok := xerr.(*exec.ExitError); ok {
			s.logger.Errorf("MSIEXEC exited with code %d", exit.ExitCode())
		}

		// pull logs
		log, err := ioutil.ReadFile(logfile.Name())
		if err != nil {
			return err
		}
		s.logger.Errorf("MSIEXEC logs: %s", string(log))

		return xerr
	}
	return nil
}

// Loader provides an instantiation function for this step plugin
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	step := &Step{
		StepName:              "",
		ProgramMatch:          "Chef Infra Client",
		MSIMatch:              "chef-client.*\\.msi",
		MSIEXECTimeoutSeconds: 300,

		logger:       go2chef.GetGlobalLogger(),
		source:       nil,
		downloadPath: "",
	}

	if err := mapstructure.Decode(config, step); err != nil {
		return nil, err
	}

	source, err := go2chef.GetSourceFromStepConfig(config)
	if err != nil {
		return nil, err
	}
	step.source = source

	return step, nil
}

func init() {
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterStep(TypeName, Loader)
	}
}

func (s *Step) findMSI() (string, error) {
	re, err := regexp.Compile(s.MSIMatch)
	if err != nil {
		return "", err
	}
	return util.MatchPath(s.downloadPath, re)
}
