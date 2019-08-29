package bundle

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

// Config defines the structure of the bundle configuration file
type Config struct {
	Shell   string   `mapstructure:"shell"`
	Command []string `mapstructure:"command"`
	Timeout int      `mapstructure:"timeout"`
}

// LoadBundleConfig loads the configuration for a bundle from a file
func LoadBundleConfig(path string) (*Config, error) {
	b := &Config{
		Shell:   "exec",
		Timeout: 1,
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, b); err != nil {
		return nil, err
	}
	return b, nil
}

// Execute executes a bundle configuration
func (bc *Config) Execute(dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(bc.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bc.Command[0], bc.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
