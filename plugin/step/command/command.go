package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

const TypeName = "go2chef.step.command"

type Step struct {
	SName          string
	Command        []string `mapstructure:"command"`
	Env            map[string]string
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
	logger         go2chef.Logger
}

func (s *Step) String() string {
	return "<" + TypeName + ":" + s.SName + ">"
}

func (s *Step) SetName(name string) {
	s.SName = name
}

func (s *Step) Name() string {
	return s.SName
}

func (s *Step) Type() string {
	return TypeName
}

func (s *Step) Download() error {
	return nil
}

func (s *Step) Execute() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.TimeoutSeconds)*time.Second)
	defer cancel()

	if len(s.Command) < 1 {
		return fmt.Errorf("empty command specification for %s", TypeName)
	}
	cmd := exec.CommandContext(ctx, s.Command[0], s.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	env := make([]string, 0, len(s.Env))
	for k, v := range s.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	return cmd.Run()
}

func Loader(config map[string]interface{}) (go2chef.Step, error) {
	c := &Step{
		logger:         go2chef.GetGlobalLogger(),
		TimeoutSeconds: 300,
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
