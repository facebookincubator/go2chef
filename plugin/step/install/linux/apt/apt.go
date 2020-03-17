package apt

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/facebookincubator/go2chef"
	"github.com/facebookincubator/go2chef/util/temp"
	"github.com/mitchellh/mapstructure"
)

// TypeNames for each variant of this step plugin
const (
	TypeName    = "go2chef.step.install.linux.apt"
	GetTypeName = "go2chef.step.install.linux.apt_get"
)

var (
	// DefaultPackageName is the default package name to use for Chef installation
	DefaultPackageName = "chef"
)

// Step implements Chef installation via Debian/Ubuntu Apt/apt-get
type Step struct {
	StepName    string `mapstructure:"name"`
	APTBinary   string `mapstructure:"apt_binary"`
	DPKGBinary  string `mapstructure:"dpkg_binary"`
	PackageName string `mapstructure:"package_name"`

	Version string `mapstructure:"version"`

	DpkgCheckTimeoutSeconds int `mapstructure:"dpkg_check_timeout_seconds"`
	InstallTimeoutSeconds   int `mapstructure:"install_timeout_seconds"`

	logger              go2chef.Logger
	source              go2chef.Source
	downloadPath        string
	packageRegex        *regexp.Regexp
	packageVersionRegex *regexp.Regexp
}

func (s *Step) String() string {
	return "<" + TypeName + ":" + s.StepName + ">"
}

// SetName sets the name of this step instance
func (s *Step) SetName(str string) {
	s.StepName = str
}

// Name gets the name of this step instance
func (s *Step) Name() string {
	return s.StepName
}

// Type returns the type of this step instance
func (s *Step) Type() string {
	return TypeName
}

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
	installPackage := s.PackageName

	if s.source != nil {
		deb, err := s.findDEB()
		if err != nil {
			s.logger.Errorf("failed to discover DEB package")
			return err
		}
		installPackage = filepath.Join(s.downloadPath, deb)
	}

	installed := false
	if s.Version != "" {
		if err := s.checkInstalled(); err != nil {
			switch err.(type) {
			case *exec.ExitError:
				s.logger.Infof("dpkg-query exited with code %d", err.(*exec.ExitError).ExitCode())
				installed = false
			case *go2chef.ErrChefAlreadyInstalled:
				s.logger.Infof("%s", err)
				installed = true
			}
		}
	}

	if !installed {
		return s.installChef(installPackage)
	}
        s.logger.Infof("%s specified is already installed, not reinstalling", installPackage)
	return nil
}

// LoaderForBinary provides an instantiation function for this step plugin specific to the passed binary
func LoaderForBinary(binary string) go2chef.StepLoader {
	return func(config map[string]interface{}) (go2chef.Step, error) {
		step := &Step{
			StepName:                "",
			APTBinary:               "/usr/bin/" + binary,
			DPKGBinary:              "/usr/bin/dpkg-query",
			PackageName:             DefaultPackageName,
			DpkgCheckTimeoutSeconds: 60,
			InstallTimeoutSeconds:   300,
			logger:                  go2chef.GetGlobalLogger(),
			source:                  nil,
			downloadPath:            "",
		}

		if err := mapstructure.Decode(config, step); err != nil {
			return nil, err
		}

		source, err := go2chef.GetSourceFromStepConfig(config)
		if err != nil {
			return nil, err
		}
		step.source = source

		reStr := fmt.Sprintf("^%s.*.deb$", step.PackageName)
		regex, err := regexp.Compile(reStr)
		if err != nil {
			step.logger.Errorf("failed to compile package matching regex %s", reStr)
			return nil, err
		}
		step.packageRegex = regex

		vreStr := fmt.Sprintf("^%s\t%s.*", step.PackageName, step.Version)
		vRegex, err := regexp.Compile(vreStr)
		if err != nil {
			step.logger.Errorf("failed to compile package version match regex %s", vreStr)
		}
		step.packageVersionRegex = vRegex

		return step, nil
	}
}

var _ go2chef.Step = &Step{}

func init() {
	go2chef.RegisterStep(TypeName, LoaderForBinary("apt"))
	go2chef.RegisterStep(GetTypeName, LoaderForBinary("apt-get"))
}

func (s *Step) findDEB() (string, error) {
	dirEntries, err := ioutil.ReadDir(s.downloadPath)
	if err != nil {
		return "", err
	}

	var matches []string
	for _, entry := range dirEntries {
		if s.packageRegex.MatchString(entry.Name()) {
			matches = append(matches, entry.Name())
		}
	}

	if len(matches) < 1 {
		return "", os.ErrNotExist
	}

	sort.Strings(matches)

	if s.Version != "" {
		for _, m := range matches {
			if strings.Contains(m, s.Version) {
				return m, nil
			}
		}
	}

	return matches[0], nil
}

func (s *Step) checkInstalled() error {
	chkCtx, chkCtxCancel := context.WithTimeout(context.Background(), time.Duration(s.DpkgCheckTimeoutSeconds)*time.Second)
	defer chkCtxCancel()

	// run `dpkg -W -f "${binary:Package}\t${Version}" chef` to get current package
	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(chkCtx, s.DPKGBinary, "-W", "-f", "${binary:Package}\t${Version}", s.PackageName)
	cmd.Stdin = nil
	cmd.Stderr = os.Stderr
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return err
	}

	inst := strings.TrimSpace(buf.String())
	if s.packageVersionRegex.MatchString(inst) {
		return &go2chef.ErrChefAlreadyInstalled{
			Installed: inst,
			Requested: s.packageVersionRegex.String(),
		}
	}

	return nil
}

func (s *Step) installChef(installPackage string) error {
	instCtx, instCtxCancel := context.WithTimeout(context.Background(), time.Duration(s.InstallTimeoutSeconds)*time.Second)
	defer instCtxCancel()

	cmd := exec.CommandContext(instCtx, s.APTBinary, "-y", "install", installPackage)
	cmd.Stdin = nil
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
