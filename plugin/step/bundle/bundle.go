package bundle

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/facebookincubator/go2chef/util/temp"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this step plugin
const TypeName = "go2chef.step.bundle"

// Bundle represents an executable bundle of files that
// can be downloaded from a go2chef.Source
type Bundle struct {
	BundleName     string `mapstructure:"name"`
	source         go2chef.Source
	logger         go2chef.Logger
	downloadPath   string
	ConfigName     string `mapstructure:"config_name"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

func (b *Bundle) String() string {
	return "<" + TypeName + ":" + b.BundleName + ">"
}

// Name returns the name of this bundle step
func (b *Bundle) Name() string {
	return b.BundleName
}

// Type returns "bundle"
func (b *Bundle) Type() string {
	return TypeName
}

// SetName sets the name of this bundle step
func (b *Bundle) SetName(n string) {
	b.BundleName = n
}

// Download fetches resources required for this bundle's execution
func (b *Bundle) Download() error {
	b.logger.Debugf(1, "%s: downloading bundle", b.Name())

	tmpdir, err := temp.Dir("", "go2chef-bundle")
	if err != nil {
		return err
	}
	if err := b.source.DownloadToPath(tmpdir); err != nil {
		return err
	}
	b.downloadPath = tmpdir
	b.logger.Debugf(1, "%s: downloaded bundle to %s", b.Name(), b.downloadPath)
	return nil
}

// Execute loads the bundle.json and executes the command specified therein
func (b *Bundle) Execute() error {
	entryPoint, err := findEntrypoint(b.downloadPath)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(b.TimeoutSeconds)*time.Second)
	defer cancel()

	var cmd *exec.Cmd

	cmd = commandForPath(entryPoint, ctx)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = b.downloadPath

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// Loader provides an instantiation function for this step
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	source, err := go2chef.GetSourceFromStepConfig(config)
	if err != nil {
		return nil, err
	}
	b := &Bundle{
		source:         source,
		logger:         go2chef.GetGlobalLogger(),
		ConfigName:     "bundle.json",
		TimeoutSeconds: 300,
	}
	if err := mapstructure.Decode(config, b); err != nil {
		return nil, err
	}
	b.source.SetName(b.Name() + "-source")

	return b, nil
}

func init() {
	go2chef.RegisterStep(TypeName, Loader)
}

var _ go2chef.Step = &Bundle{}
var _ go2chef.StepLoader = Loader
