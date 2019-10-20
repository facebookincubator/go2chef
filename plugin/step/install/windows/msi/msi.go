package msi

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/facebookincubator/go2chef/util/temp"

	"github.com/facebookincubator/go2chef/util"

	"github.com/mitchellh/mapstructure"

	"github.com/facebookincubator/go2chef"

	"golang.org/x/sys/windows/registry"
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

	Cleanup   bool `mapstructure:"cleanup"`
	Uninstall bool `mapstructure:"uninstall"`

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

	if running, err := isMSIRunning(); running || err != nil {
		if err != nil {
			return err
		}

		return errors.New(`another msi is installing`)
	}

	if s.Uninstall {
		if err = uninstallChef(s.MSIEXECTimeoutSeconds); err != nil {
			s.logger.Debugf(1, "%s", err)
			// Try to install anyway I guess...
		}
	}

	if s.Cleanup {
		cleanupChefDirectory()
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

func isMSIRunning() (bool, error) {
	cmd := exec.Command("sc.exe", "query", "msiserver")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	re := regexp.MustCompile(`\s+STATE\s+: \d+\s+([a-zA-Z]+)`)
	m := re.FindAllSubmatch(out, -1)
	state := string(m[0][1])
	if state != "STOPPED" {
		return true, nil
	}

	return false, nil
}

type chefInstallInfo struct {
	Path          string
	UninstallGUID string
	Version       string
	Installed     bool
}

const installedProducts = `SOFTWARE\Classes\Installer\Products`

func testChefInstalled() (*chefInstallInfo, error) {
	re := regexp.MustCompile(`Chef (Infra ){0,1}Client v([\d\.]+)\s*`)
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, installedProducts, registry.QUERY_VALUE|registry.READ)
	if err != nil {
		return &chefInstallInfo{Installed: false}, err
	}
	defer k.Close()

	ks, err := k.Stat()
	if err != nil {
		return &chefInstallInfo{Installed: false}, err
	}

	kn, err := k.ReadSubKeyNames(int(ks.SubKeyCount))
	if err != nil {
		return &chefInstallInfo{Installed: false}, err
	}

	for _, s := range kn {
		searchKey := strings.Join([]string{installedProducts, s}, `\`)
		searchSubKey, err := registry.OpenKey(registry.LOCAL_MACHINE, searchKey, registry.QUERY_VALUE|registry.READ)
		if err != nil {
			continue
		}
		defer searchSubKey.Close()

		pn, _, err := searchSubKey.GetStringValue("ProductName")
		if err != nil {
			continue
		}

		if re.MatchString(pn) {
			var uninstallGUID string

			pi, _, err := searchSubKey.GetStringValue("ProductIcon")
			if err == nil {
				for _, c := range strings.Split(pi, `\`) {
					if strings.Contains(c, `{`) {
						uninstallGUID = c
						break
					}
				}
			}

			verMatch := re.FindAllStringSubmatch(pn, -1)[0][2]

			return &chefInstallInfo{
				Installed:     true,
				Path:          searchKey,
				UninstallGUID: uninstallGUID,
				Version:       verMatch,
			}, nil
		}
	}

	return &chefInstallInfo{Installed: false}, nil
}

func cleanupChefDirectory() error {
	chefInstallDir := `C:\opscode\chef`

	if info, _ := os.Stat(chefInstallDir); info == nil {
		return nil
	}

	var (
		recycleBin  = `C:\$Recycle.Bin`
		done        = make(chan struct{})
		ctx, cancel = context.WithTimeout(context.Background(), time.Minute*10)
	)

	go func() {
		if trash, err := ioutil.TempDir(recycleBin, "go2chef"); err == nil {
			exec.CommandContext(ctx, "move", "/Y", chefInstallDir, trash)
			done <- struct{}{}
			return
		}

		cancel()
	}()

	select {
	case <-done:
		fmt.Println("finished a okay!")
	case <-ctx.Done():
		return errors.New("cancelled")
	}

	return nil
}

func uninstallChef(timeout int) error {
	var (
		chefInfo *chefInstallInfo
		err      error
	)

	if chefInfo, err = testChefInstalled(); err != nil {
		return err
	}

	done := make(chan error)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout))
	defer cancel()

	go func() {
		cmd := exec.CommandContext(ctx, "msiexec", "/qn", "/x", chefInfo.UninstallGUID)
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return errors.New(`uninstall timed out`)
	}

	return nil
}
