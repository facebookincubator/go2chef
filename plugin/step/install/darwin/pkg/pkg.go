package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/facebookincubator/go2chef/util"

	"github.com/mitchellh/mapstructure"

	"github.com/facebookincubator/go2chef"
)

const TypeName = "go2chef.step.install.darwin.pkg"

type Step struct {
	// StepName defines the name of the step
	StepName string `mapstructure:"name"`
	// PKGMatch specifies a regex to match the filenames in this step's
	// source directory against to find the PKG file
	PKGMatch string `mapstructure:"pkg_match"`
	// DMGMatch specifies a regex to match the filenames in this step's
	// source directory against to find the DMG file
	DMGMatch string `mapstructure:"dmg_match"`
	// InstallerTimeoutSeconds defines the timeout for installation
	InstallerTimeoutSeconds int `mapstructure:"installer_timeout_seconds"`
	// IsDMG enables installation of a pkg inside a DMG
	IsDMG bool `mapstructure:"is_dmg"`

	logger       go2chef.Logger
	source       go2chef.Source
	downloadPath string
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
	// If this is a DMG, go down the rabbit hole. Mount it and
	// then set downloadPath to its mount point.
	if s.IsDMG {
		s.logger.WriteEvent(go2chef.NewEvent("INSTALL_PKG_DMG_MOUNT", s.Name(), "mounting DMG"))
		dmg, err := s.findDMG()
		if err != nil {
			return err
		}
		if err := s.mountDMG(filepath.Join(s.downloadPath, dmg)); err != nil {
			return err
		}

		// unmount and emit events
		defer func() {
			if err := s.unmountDMG(); err != nil {
				s.logger.WriteEvent(go2chef.NewEvent("INSTALL_PKG_DMG_UNMOUNT_FAILED", s.Name(), "unmounting DMG failed!"))

			}
			s.logger.WriteEvent(go2chef.NewEvent("INSTALL_PKG_DMG_UNMOUNT", s.Name(), "unmounting DMG"))
		}()
	}

	pkg, err := s.findPKG()
	if err != nil {
		return err
	}

	instCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.InstallerTimeoutSeconds)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(instCtx, "installer", "-verbose", "-pkg", filepath.Join(s.downloadPath, pkg), "-target", "/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// preserve exit error
		xerr := err
		if exit, ok := xerr.(*exec.ExitError); ok {
			s.logger.Errorf("pkg installer exited with code %d", exit.ExitCode())
		}
		return xerr
	}
	return nil
}

func Loader(config map[string]interface{}) (go2chef.Step, error) {
	step := &Step{
		StepName:                "",
		PKGMatch:                "chef.*",
		DMGMatch:                "chef.*",
		InstallerTimeoutSeconds: 300,
		IsDMG:                   false,

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

func (s *Step) findPKG() (string, error) {
	re, err := regexp.Compile(s.PKGMatch)
	if err != nil {
		return "", err
	}
	return util.MatchPath(s.downloadPath, re)
}

func (s *Step) findDMG() (string, error) {
	re, err := regexp.Compile(s.DMGMatch)
	if err != nil {
		return "", err
	}
	return util.MatchPath(s.downloadPath, re)
}

func (s *Step) mountDMG(dmg string) error {
	ctx := context.Background()

	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx,
		"hdiutil", "mount", dmg,
		"-mountroot", tmpdir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	dir, err := getDMGVolume(tmpdir)
	if err != nil {
		s.downloadPath = tmpdir
		defer s.unmountDMG()
		return err
	}

	s.downloadPath = dir
	return nil
}

func (s *Step) unmountDMG() error {
	ctx := context.Background()

	cmd := exec.CommandContext(ctx, "hdiutil", "detach", s.downloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getDMGVolume(dir string) (string, error) {
	ents, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	if len(ents) < 1 {
		return "", fmt.Errorf("DMG volume mount contains no volume folder")
	}
	return filepath.Join(dir, ents[0].Name()), nil
}
