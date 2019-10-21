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

	RenameFolder bool `mapstructure:"rename_folder"`
	Uninstall    bool `mapstructure:"uninstall"`

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
func (s *Step) Execute() (err error) {
	if running, err := s.isMSIRunning(); running || err != nil {
		if err != nil {
			return err
		}

		return errors.New(`another msi is installing`)
	}

	if s.Uninstall {
		if err = s.uninstallChef(s.MSIEXECTimeoutSeconds); err != nil {
			s.logger.Debugf(1, "%s", err)
			return err
		}
	}

	if s.RenameFolder {
		if err = s.renameFolder(s.MSIEXECTimeoutSeconds); err != nil {
			s.logger.Debugf(1, "%s", err)
			return err
		}
	}

	return s.installChef()
}

func (s *Step) installChef() error {
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

// A simple check to see if the msiserver has a lock, which prevents installation
// of other MSIs.
// It's possible for this service to run for a while after an msi has terminated,
// but there is no point in trying when it won't even work  ¯\_(ツ)_/¯
func (s *Step) isMSIRunning() (bool, error) {
	cmd := exec.Command("sc.exe", "query", "msiserver")
	out, err := cmd.Output()
	if err != nil {
		s.logger.Errorf("%s", err)
		return false, err
	}

	re := regexp.MustCompile(`\s+STATE\s+: \d+\s+([a-zA-Z]+)`)
	m := re.FindAllSubmatch(out, -1)
	state := string(m[0][1])
	if state != "STOPPED" {
		s.logger.Debugf(2, "msiserver is not stopped")
		return true, nil
	}

	s.logger.Debugf(2, "msiserver is stopped")
	return false, nil
}

// The MSI of the installation is recorded in the registry. We can use this
// information to check if the desired version of Chef is already installed.
type chefInstallInfo struct {
	Path          string
	UninstallGUID string
	Version       string
	Installed     bool
}

const (
	installedProducts       = `SOFTWARE\Classes\Installer\Products`
	registryReadPermissions = registry.QUERY_VALUE | registry.READ
)

// Scans the registry for installed products. It will find a product name that
// matches a regex which contains enough information about what is installed.
// The resulting struct can be used to uninstall the old client, if desired, or
// make a judgement call of if a new versions has to be installed.
func (s *Step) testChefInstalled() (*chefInstallInfo, error) {
	re := regexp.MustCompile(`Chef (Infra ){0,1}Client v([\d\.]+)\s*`)
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, installedProducts, registryReadPermissions)
	if err != nil {
		s.logger.Errorf("%s", err)
		return &chefInstallInfo{Installed: false}, err
	}
	defer k.Close()

	ks, err := k.Stat()
	if err != nil {
		s.logger.Errorf("%s", err)
		return &chefInstallInfo{Installed: false}, err
	}

	kn, err := k.ReadSubKeyNames(int(ks.SubKeyCount))
	if err != nil {
		s.logger.Errorf("%s", err)
		return &chefInstallInfo{Installed: false}, err
	}

	for _, s := range kn {
		searchKey := strings.Join([]string{installedProducts, s}, `\`)
		searchSubKey, err := registry.OpenKey(registry.LOCAL_MACHINE, searchKey, registryReadPermissions)
		if err != nil {
			continue
		}
		defer searchSubKey.Close()

		pn, _, err := searchSubKey.GetStringValue("ProductName")
		if err != nil {
			continue
		}

		if re.MatchString(pn) {
			result := &chefInstallInfo{
				Installed: true,
				Path:      searchKey,
			}

			// This contains a path on disk to the product icon.
			// From here we can infer the application's GUID.
			// Too bad this information doesn't appear to be stored directly in the
			// registry =(
			pi, _, err := searchSubKey.GetStringValue("ProductIcon")
			if err == nil {
				for _, c := range strings.Split(pi, `\`) {
					if strings.Contains(c, `{`) {
						result.UninstallGUID = c
						break
					}
				}
			}

			verMatch := re.FindAllStringSubmatch(pn, -1)
			if len(verMatch) > 0 && len(verMatch[0]) > 2 {
				result.Version = verMatch[0][2]
			}

			return result, nil
		}
	}

	return &chefInstallInfo{Installed: false}, nil
}

/*
	Sometimes there is a file within the Chef directory that has a lock on a file.
	The installer will fail to remove this file. In this case when an installation
	attempt is made it could fail to finish and then the new client won't be installed.

	Congratulations!

	Now your node is in an inconsistent state! If you're relying on
	Chef to recover from this it is fairly challenging since the application won't run
	and yet Windows will still think it's installed correctly.

	Instead of relying on the MSI installation/upgrade to work correctly (it hasn't
	since the early versions of Chef 12), move the old installation directory out
	of the way. The installation will now be able to successfully complete since
	there are no locked files to contend with!
*/
func (s *Step) renameFolder(timeout int) (err error) {
	const (
		chefInstallDir = `C:\opscode\chef`
		recycleBin     = `C:\$Recycle.Bin`
	)
	var trash string

	if info, _ := os.Stat(chefInstallDir); info == nil {
		return nil
	}

	if trash, err = ioutil.TempDir(recycleBin, "go2chef"); err != nil {
		return fmt.Errorf("could not create temporary directory: %s", err)
	}

	if err := os.Rename(chefInstallDir, trash); err != nil {
		return err
	}

	return nil
}

// Use the information collected from the registry to uninstall the client.
func (s *Step) uninstallChef(timeout int) error {
	var (
		chefInfo *chefInstallInfo
		err      error
	)

	if chefInfo, err = s.testChefInstalled(); err != nil {
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
