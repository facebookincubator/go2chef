package command

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/facebookincubator/go2chef"
	"github.com/facebookincubator/go2chef/util/temp"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this step plugin
const TypeName = "go2chef.step.command"

// Step implements a command execution step plugin
type Step struct {
	SName          string   `mapstructure:"name"`
	Command        []string `mapstructure:"command"`
	Env            map[string]string
	TimeoutSeconds int      `mapstructure:"timeout_seconds"`
	PassthroughEnv []string `mapstructure:"passthrough_env"`

	source       go2chef.Source
	Output       map[string]string `mapstructure:"output"`
	logger       go2chef.Logger
	downloadPath string
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
	if s.source == nil {
		return nil
	}
	s.logger.Debugf(1, "%s: downloading source", s.Name())

	tmpdir, err := temp.Dir("", "go2chef-bundle")
	if err != nil {
		return err
	}
	if err := s.source.DownloadToPath(tmpdir); err != nil {
		return err
	}
	s.downloadPath = tmpdir
	s.logger.Debugf(1, "%s: downloaded source to %s", s.Name(), s.downloadPath)
	return nil
}

// Execute runs the actual command.
func (s *Step) Execute() error {
	var err error
	var outFile *os.File
	var errFile *os.File
	ctx := context.Background()
	if s.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s.TimeoutSeconds)*time.Second)
		defer cancel()
	}
	if len(s.Command) < 1 {
		return fmt.Errorf("empty command specification for %s", TypeName)
	}

	cmd := exec.CommandContext(ctx, s.Command[0], s.Command[1:]...)
	cmd, outFile, errFile, err = setOutputRedirect(s.Output, cmd)
	if outFile != nil {
		defer outFile.Close()
	}
	if errFile != nil {
		defer errFile.Close()
	}
	if err != nil {
		return err
	}
	cmd.Dir = s.downloadPath
	env := make([]string, 0, len(s.Env))
	for _, ev := range s.PassthroughEnv {
		for _, eval := range os.Environ() {
			if strings.HasPrefix(eval, ev) {
				env = append(env, eval)
			}
		}
	}
	for k, v := range s.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	return cmd.Run()
}

func setOutputRedirect(output map[string]string, cmd *exec.Cmd) (*exec.Cmd, *os.File, *os.File, error) {
	var mw io.Writer
	var outFile *os.File
	var errFile *os.File
	var err error
	filePath := output["out"]
	errFilePath := output["err"]
	if filePath != "" {
		mw, outFile, err = setOutputWrite(filePath)
		if err != nil {
			return nil, nil, nil, err
		}
		cmd.Stdout = mw
		if errFilePath == filePath {
			cmd.Stderr = cmd.Stdout
		}
	} else {
		outFile = nil
		cmd.Stdout = os.Stdout
	}
	if filePath != errFilePath && errFilePath != "" {
		mw, errFile, err = setOutputWrite(errFilePath)
		if err != nil {
			return nil, nil, nil, err
		}
		cmd.Stderr = mw
	} else {
		errFile = nil
		cmd.Stderr = os.Stderr
	}
	return cmd, outFile, errFile, nil
}

func setOutputWrite(path string) (io.Writer, *os.File, error) {
	var file *os.File
	if _, err := os.Stat(path); err != nil && errors.Is(err, fs.ErrNotExist) {
		file, err = create(path)
		if err != nil {
			return nil, nil, err
		}
	} else {
		file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, nil, err
		}
	}
	return io.MultiWriter(file, os.Stdout), file, nil
}

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

// Loader provides an instantiation function for this step plugin
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	source, err := go2chef.GetSourceFromStepConfig(config)
	if err != nil {
		return nil, err
	}
	c := &Step{
		logger:         go2chef.GetGlobalLogger(),
		TimeoutSeconds: 0,
		Command:        make([]string, 0),
		Env:            make(map[string]string),
		source:         source,
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
