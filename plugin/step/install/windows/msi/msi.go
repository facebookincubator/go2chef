package msi

import (
	"bytes"
	"context"
	"encoding/json"
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

const TypeName = "go2chef.step.install.windows.msi"

var (
	DefaultPackageName = "chef"
)

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

func (s *Step) SetName(str string) { s.StepName = str }

func (s *Step) Name() string { return s.StepName }

func (s *Step) Type() string { return TypeName }

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
	msi, err := s.findMSI()
	if err != nil {
		return err
	}

	instCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.MSIEXECTimeoutSeconds)*time.Second)
	defer cancel()

	// create a logfile for MSIEXEC
	logfile, err := ioutil.TempFile("", "")
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
	return util.MatchFile(s.downloadPath, re)
}

func getInstalledPrograms() ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.CommandContext(
		context.Background(),
		"powershell.exe",
		"-Command",
		`Get-ChildItem "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall" | % { Get-ItemProperty $_.PSPath -Name DisplayName | Select -Property DisplayName } | Sort-Object -Property DisplayName | ConvertTo-Json`,
	)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data := make([]struct {
		DisplayName string
	}, 0)

	if err := json.Unmarshal(buf.Bytes(), data); err != nil {
		return nil, err
	}

	output := make([]string, len(data))
	for _, d := range data {
		output = append(output, d.DisplayName)
	}
	return output, nil
}

func isChefInstalled(program, version string) (bool, error) {
	re, err := regexp.Compile(
		fmt.Sprintf("^%s %s$", program, version),
	)
	if err != nil {
		return false, err
	}

	progs, err := getInstalledPrograms()
	if err != nil {
		return false, err
	}
	for _, prog := range progs {
		if re.MatchString(prog) {
			return true, nil
		}
	}

	return false, nil
}
