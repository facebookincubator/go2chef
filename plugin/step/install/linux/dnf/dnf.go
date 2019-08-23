/*
	Package dnf implements a step plugin for installation of Chef on RPM-based systems.
	It supports DNF, Yum, and direct RPM installation.

	If you provide a `source` config block, this plugin will download it and search for
	an RPM based on `package_name` (and, if specified, `version`).

	Example config for a Chef install

		{
			"type": "go2chef.step.install.linux.dnf",
			"name": "install chef",
			"package_name": "chef"
		}
*/
package dnf

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/facebookincubator/go2chef/util"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

const (
	TypeName    = "go2chef.step.install.linux.dnf"
	YumTypeName = "go2chef.step.install.linux.yum"
	RPMTypeName = "go2chef.step.install.linux.rpm"
)

var (
	DefaultPackageName = "chef"
)

type Step struct {
	StepName    string `mapstructure:"name"`
	DNFBinary   string `mapstructure:"dnf_binary"`
	RPMBinary   string `mapstructure:"rpm_binary"`
	PackageName string `mapstructure:"package_name"`

	Version string `mapstructure:"version"`

	RPMCheckTimeoutSeconds int `mapstructure:"rpm_check_timeout_seconds"`
	InstallTimeoutSeconds  int `mapstructure:"install_timeout_seconds"`

	installWithRPM bool

	logger              go2chef.Logger
	source              go2chef.Source
	downloadPath        string
	packageRegex        *regexp.Regexp
	packageVersionRegex *regexp.Regexp
}

func (s *Step) String() string {
	return "<" + TypeName + ":" + s.StepName + ">"
}

func (s *Step) SetName(str string) {
	s.StepName = str
}

func (s *Step) Name() string {
	return s.StepName
}

func (s *Step) Type() string {
	return TypeName
}

func (s *Step) Download() error {
	if s.source == nil {
		return nil
	}

	tmpdir, err := ioutil.TempDir("", "go2chef-install")
	if err != nil {
		return err
	}

	if err := s.source.DownloadToPath(tmpdir); err != nil {
		return err
	}
	s.downloadPath = tmpdir

	return nil
}

func (s *Step) Execute() error {
	installPackage := s.PackageName

	if s.source != nil {
		rpm, err := s.findRPM()
		if err != nil {
			return err
		}
		installPackage = filepath.Join(s.downloadPath, rpm)
	}

	installed := false
	if s.Version != "" {
		if err := s.checkInstalled(); err != nil {
			switch err.(type) {
			case *exec.ExitError:
				installed = false
			case *go2chef.ErrChefAlreadyInstalled:
				s.logger.Warningf("%s", err)
				installed = true
			}
		}
	}

	if !installed {
		if s.installWithRPM {
			return s.installChefRPM(installPackage)
		}
		return s.installChefDNF(installPackage)
	}
	s.logger.Warningf("Chef version specified is already installed, not reinstalling")
	return nil
}

func LoaderForBinary(binary string) go2chef.StepLoader {
	return func(config map[string]interface{}) (go2chef.Step, error) {
		step := &Step{
			StepName:               "",
			DNFBinary:              "/usr/bin/" + binary,
			RPMBinary:              "/usr/bin/rpm",
			PackageName:            DefaultPackageName,
			RPMCheckTimeoutSeconds: 60,
			InstallTimeoutSeconds:  300,

			installWithRPM: binary == "rpm",

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

		reStr := fmt.Sprintf("^%s-.*.rpm$", step.PackageName)
		regex, err := regexp.Compile(reStr)
		if err != nil {
			step.logger.Errorf("failed to compile package matching regex %s", reStr)
			return nil, err
		}
		step.packageRegex = regex

		vreStr := fmt.Sprintf("^%s-%s.*", step.PackageName, step.Version)
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
	go2chef.RegisterStep(TypeName, LoaderForBinary("dnf"))
	go2chef.RegisterStep(YumTypeName, LoaderForBinary("yum"))
	go2chef.RegisterStep(RPMTypeName, LoaderForBinary("rpm"))
}

func (s *Step) findRPM() (string, error) {
	return util.MatchFile(s.downloadPath, s.packageRegex)
}

func (s *Step) checkInstalled() error {
	chkCtx, chkCtxCancel := context.WithTimeout(context.Background(), time.Duration(s.RPMCheckTimeoutSeconds)*time.Second)
	defer chkCtxCancel()

	// run rpm -q <package> to get current package
	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(chkCtx, s.RPMBinary, "-q", s.PackageName)
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

func (s *Step) installChefDNF(installPackage string) error {
	instCtx, instCtxCancel := context.WithTimeout(context.Background(), time.Duration(s.InstallTimeoutSeconds)*time.Second)
	defer instCtxCancel()

	cmd := exec.CommandContext(instCtx, s.DNFBinary, "-y", "install", installPackage)
	cmd.Stdin = nil
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func (s *Step) installChefRPM(installPackage string) error {
	instCtx, instCtxCancel := context.WithTimeout(context.Background(), time.Duration(s.InstallTimeoutSeconds)*time.Second)
	defer instCtxCancel()

	cmd := exec.CommandContext(instCtx, s.RPMBinary, "-Uvh", installPackage)
	cmd.Stdin = nil
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}