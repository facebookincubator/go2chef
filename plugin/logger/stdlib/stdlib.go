package stdlib

import (
	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
	"github.com/oko/logif"
)

const TypeName = "go2chef.logger.stdlib"

// Logger represents a logger that just sends output to the
// default stdlib log library with some level info.
type Logger struct {
	*logif.StdlibLogger
	LoggerName string
}

// NewFromExistingStdlibLogger creates a new StdlibLogger from an existing
// logif.StdlibLogger (for use during the pre-config-loading phase)
func NewFromExistingStdlibLogger(logger *logif.StdlibLogger) *Logger {
	return &Logger{
		logger,
		"default-go2chef-cli",
	}
}

func (l *Logger) String() string {
	return "<" + TypeName + ":" + l.LoggerName + ">"
}

// Name returns the name of this logger instance
func (l *Logger) Name() string { return l.LoggerName }

// SetName sets the name of this logger instance
func (l *Logger) SetName(s string) { l.LoggerName = s }

// Type returns the type of this logger instance
func (l *Logger) Type() string {
	return TypeName
}

// WriteEvent writes a formatted event at INFO level
func (l *Logger) WriteEvent(e *go2chef.Event) {
	l.Infof("EVENT: %s in %s - %s", e.Event, e.Component, e.Message)
}

// Loader creates an StdlibLogger from a config map
func Loader(config map[string]interface{}) (go2chef.Logger, error) {
	name, _, err := go2chef.GetNameType(config)
	if err != nil {
		return nil, err
	}
	ret := &Logger{
		&logif.StdlibLogger{},
		name,
	}
	parse := struct {
		Level     string
		Debugging int
		Verbosity int
	}{}
	if err := mapstructure.Decode(config, &parse); err != nil {
		return nil, err
	}
	realLevel, err := logif.ParseLogLevel(parse.Level)
	if err != nil {
		return nil, err
	}

	// set all levels based on config
	ret.SetLevel(realLevel)
	ret.SetDebugging(parse.Debugging)
	ret.SetVerbosity(parse.Verbosity)

	return ret, nil
}

// Shutdown is a no-op for this logger
func (l *Logger) Shutdown() {}

var _ go2chef.Logger = &Logger{}
var _ go2chef.LoggerLoader = Loader

func init() {
	if go2chef.AutoRegisterPlugins {
		go2chef.RegisterLogger(TypeName, Loader)
	}
}
