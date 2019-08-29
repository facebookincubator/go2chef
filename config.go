package go2chef

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
)

// ConfigSource defines the interface for configuration sources. These can be implemented however you like.
type ConfigSource interface {
	// InitFlags sets up command line flags
	InitFlags(set *pflag.FlagSet)
	// ReadConfig actually reads in the configuration
	ReadConfig() (map[string]interface{}, error)
}

var configSourceRegistry = make(map[string]ConfigSource)

// RegisterConfigSource registers a new configuration source plugin
func RegisterConfigSource(name string, cs ConfigSource) {
	if _, ok := configSourceRegistry[name]; ok {
		panic("ConfigSource " + name + " is already registered")
	}
	configSourceRegistry[name] = cs
}

// InitializeConfigSourceFlags initializes a flag set with the flags required by
// the currently registered configuration source plugins.
func InitializeConfigSourceFlags(set *pflag.FlagSet) {
	for _, cs := range configSourceRegistry {
		cs.InitFlags(set)
	}
}

// GetConfigSource gets a specified configuration source plugin
func GetConfigSource(name string) ConfigSource {
	if cs, ok := configSourceRegistry[name]; ok {
		return cs
	}
	return nil
}

// Config defines the configuration for all of go2chef
type Config struct {
	Loggers []Logger
	Steps   []Step
}

// Global is the global configuration store
var Global = &GlobalConfig{}

// GetConfig loads and resolves the configuration
func GetConfig(configSourceName string, earlyLogger Logger) (*Config, error) {
	EarlyLogger.Printf("loading config from source %s", configSourceName)

	// Get the chosen configuration source and read
	configSource := GetConfigSource(configSourceName)
	if configSource == nil {
		return nil, &ErrComponentDoesNotExist{Component: "ConfigSource::" + configSourceName}
	}
	config, err := configSource.ReadConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	// pull global configuration
	global, err := GetGlobalConfig(config)
	if err != nil {
		return nil, err
	}
	Global = global

	loggers, err := GetLoggers(config)
	if err != nil {
		return nil, err
	}

	// if we're provided a copy of the early logger used
	// before config load, then add it to the logger list
	if earlyLogger != nil {
		realLoggers := make([]Logger, len(loggers)+1)
		copy(realLoggers[1:], loggers)
		realLoggers[0] = earlyLogger
		loggers = realLoggers
	}

	cfg.Loggers = loggers

	// initialize the global logger!
	InitGlobalLogger(cfg.Loggers)

	// pull steps
	steps, err := GetSteps(config)
	if err != nil {
		return nil, err
	}
	cfg.Steps = steps

	return cfg, nil
}

// GetLoggers extracts an array of loggers from a config map
func GetLoggers(config map[string]interface{}) ([]Logger, error) {
	parse := struct {
		Loggers []map[string]interface{} `mapstructure:"loggers"`
	}{}
	if err := mapstructure.Decode(config, &parse); err != nil {
		return nil, err
	}
	loggers := make([]Logger, 0, len(parse.Loggers))
	for _, lconf := range parse.Loggers {
		_, ltype, err := GetNameType(lconf)
		if err != nil {
			return nil, err
		}

		logger, err := GetLogger(ltype, lconf)
		if err != nil {
			return nil, err
		}

		loggers = append(loggers, logger)
	}
	return loggers, nil
}

// GetSteps extracts an array of steps from a config map
func GetSteps(config map[string]interface{}) ([]Step, error) {
	parse := struct {
		Steps []map[string]interface{} `mapstructure:"steps"`
	}{}

	if err := mapstructure.Decode(config, &parse); err != nil {
		return nil, err
	}
	steps := make([]Step, 0, len(parse.Steps))
	for _, sconf := range parse.Steps {
		_, stype, err := GetNameType(sconf)
		if err != nil {
			return nil, err
		}

		step, err := GetStep(stype, sconf)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}
	return steps, nil
}

// GetSourceFromStepConfig gets a Source from a Step's config map. If there is
// no `source` key, then it will return a nil Source and no error.
func GetSourceFromStepConfig(config map[string]interface{}) (Source, error) {
	parse := struct {
		Source map[string]interface{} `mapstructure:"source"`
	}{}

	if err := mapstructure.Decode(config, &parse); err != nil {
		return nil, err
	}

	if parse.Source == nil {
		return nil, nil
	}
	if len(parse.Source) == 0 {
		return nil, nil
	}

	stype, err := GetType(parse.Source)
	if err != nil {
		return nil, err
	}

	src, err := GetSource(stype, parse.Source)
	if err != nil {
		return nil, err
	}
	return src, nil
}
