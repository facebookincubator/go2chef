package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this step plugin
const TypeName = "go2chef.step.command"

// Step implements a command execution step plugin
type Step struct {
	SName          string
	Command        []string `mapstructure:"command"`
	Env            map[string]string
	TimeoutSeconds int      `mapstructure:"timeout_seconds"`
	PassthroughEnv []string `mapstructure:"passthrough_env"`
	logger         go2chef.Logger
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

// Execute runs the actual command.
func (s *Step) Execute() error {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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

// Loader provides an instantiation function for this step plugin
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	c := &Step{
		logger:         go2chef.GetGlobalLogger(),
		TimeoutSeconds: 0,
		Command:        make([]string, 0),
		Env:            make(map[string]string),
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
