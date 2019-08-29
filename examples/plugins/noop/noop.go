package noop

import (
	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of the noop plugin
const TypeName = "go2chef.examples.plugin.noop"

// Step is where you should define the internal data structure
// of your step. For simple steps, you can directly unpack the
// config map[string]interface{} directly onto this struct, or
// for more complex designs you might want to unpack to an
// options struct first for validation/processing.
type Step struct {
	SName  string `mapstructure:"name"`
	logger go2chef.Logger
}

func (s *Step) String() string { return "<" + TypeName + ":" + s.SName + ">" }

// Name returns the name of this noop step
func (s *Step) Name() string { return s.SName }

// Type returns the type of this noop step, "noop"
func (s *Step) Type() string { return "noop" }

// SetName allows parent components to set the step name
func (s *Step) SetName(n string) {}

// Download is where you should perform resource retrieval and
// validation. Downloads for step groups are executed in
// parallel, so don't do any system modifications in Download().
func (s *Step) Download() error {
	s.logger.Infof("noop %s: Download() called", s.Name())
	return nil
}

// Execute is where you should actually perform system actions
// in your plugin. Execute calls are performed serially within
// step groups.
func (s *Step) Execute() error {
	s.logger.Infof("noop %s: Execute() called", s.Name())
	return nil
}

// Loader is the function that takes a config and generates a fully
// configured plugin instance. This function is what's placed in the
// plugin registry and called for each config that requests a `noop`
// step type.
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	// Here, we'll create a sane default instance of the plugin, and
	// then customize it using the config map passed in from the parent.
	// For logging, just pull go2chef.GetGlobalLogger()
	s := &Step{
		SName:  "noop",
		logger: go2chef.GetGlobalLogger(),
	}
	if err := mapstructure.Decode(config, s); err != nil {
		return nil, err
	}
	return s, nil
}

// I find it helpful to have these type checking no-ops for IDE cues,
// but they aren't strictly necessary for anything.
var _ go2chef.Step = &Step{}
var _ go2chef.StepLoader = Loader

func init() {
	// You have to register your step init function in order
	// for it to be usable, and you have to do that before
	// Go2Chef loads, so put it in your module's init().
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterStep("noop", Loader)
	}
}
