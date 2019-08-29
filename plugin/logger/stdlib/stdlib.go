package stdlib

import (
	"log"
	"os"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

const TypeName = "go2chef.logger.stdlib"

// Logger represents a logger that just sends output to the
// default stdlib log library with some level info.
type Logger struct {
	LoggerName string
	log        *log.Logger
	level      int
	debug      int
}

type Config struct {
	Level     string
	Debugging int
}

func NewFromLogger(l *log.Logger, level, debug int) *Logger {
	return &Logger{
		LoggerName: "go2chef",
		log:        l,
		level:      level,
		debug:      debug,
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

func (l *Logger) SetDebug(dbg int) {
	l.debug = dbg
}
func (l *Logger) SetLevel(lvl int) {
	l.level = lvl
}

func (l *Logger) Errorf(fmt string, args ...interface{}) {
	l.log.Printf("ERROR: "+fmt, args...)
}

func (l *Logger) Infof(fmt string, args ...interface{}) {
	if l.level >= go2chef.LogLevelInfo {
		l.log.Printf("INFO: "+fmt, args...)
	}
}

func (l *Logger) Debugf(dbg int, fmt string, args ...interface{}) {
	if l.level >= go2chef.LogLevelDebug && dbg >= l.debug {
		l.log.Printf("DEBUG: "+fmt, args...)
	}
}

// WriteEvent writes a formatted event at INFO level
func (l *Logger) WriteEvent(e *go2chef.Event) {
	l.log.Printf("EVENT: %s in %s - %s", e.Event, e.Component, e.Message)
}

// Loader creates an StdlibLogger from a config map
func Loader(config map[string]interface{}) (go2chef.Logger, error) {
	name, _, err := go2chef.GetNameType(config)
	if err != nil {
		return nil, err
	}
	parse := Config{}
	ret := &Logger{
		name,
		log.New(os.Stderr, "GO2CHEF", log.LstdFlags),
		go2chef.LogLevelInfo,
		0,
	}
	if err := mapstructure.Decode(config, &parse); err != nil {
		return nil, err
	}
	realLevel, err := go2chef.StringToLogLevel(parse.Level)
	if err != nil {
		return nil, err
	}

	// set all levels based on config
	ret.SetLevel(realLevel)
	ret.SetDebug(parse.Debugging)

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
