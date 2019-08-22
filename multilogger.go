package go2chef

import (
	"math"

	"github.com/oko/logif"
)

// MultiLogger is a fan-out logger for use as the central
// logging broker in go2chef
type MultiLogger struct {
	loggers   []Logger
	verbosity int
	debug     int
	parent    *MultiLogger
	level     int
}

// NewMultiLogger returns a MultiLogger with the provided list
// of loggers set up to receive logs.
func NewMultiLogger(loggers []Logger) *MultiLogger {
	return &MultiLogger{
		loggers:   loggers,
		verbosity: math.MaxInt64,
		debug:     math.MaxInt64,
		parent:    nil,
		level:     logif.LevelDebug,
	}
}

func (m *MultiLogger) String() string {
	return "MultiLogger"
}

// Name returns the name of this logger
func (m *MultiLogger) Name() string {
	return "MultiLogger"
}

// SetName is a no-op for this logger
func (m *MultiLogger) SetName(string) {}

// Type returns the type of this logger
func (m *MultiLogger) Type() string {
	return "go2chef.logger.multi"
}

// Errorf logs a formatted message at ERROR level
func (m *MultiLogger) Errorf(s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Errorf(s, v...)
	}
}

// Warningf logs a formatted message at WARN level
func (m *MultiLogger) Warningf(s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Warningf(s, v...)
	}
}

// Infof logs a formatted message at INFO level
func (m *MultiLogger) Infof(s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Infof(s, v...)
	}
}

// Debugf logs a formatted message at DEBUG level
func (m *MultiLogger) Debugf(s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Debugf(s, v...)
	}
}

// V returns a sublogger that only logs INFO messages if the parent verbosity is greater than v
func (m *MultiLogger) V(v int) logif.Logger {
	return m
}

// D returns a sublogger that only logs DEBUG messages if the parent debugging verbosity is greater than d
func (m *MultiLogger) D(d int) logif.Logger {
	return m
}

// IsV returns whether verbosity is set at the given level
func (m *MultiLogger) IsV(v int) bool {
	return true
}

// IsD returns whether debugging is set at the given level
func (m *MultiLogger) IsD(d int) bool {
	return true
}

// Verbosity returns the logger verbosity
func (m *MultiLogger) Verbosity() int {
	return m.verbosity
}

// Debugging returns the logger debugging verbosity
func (m *MultiLogger) Debugging() int {
	return m.debug
}

// Level gets the logger's overall level threshold
func (m *MultiLogger) Level() int {
	return m.level
}

// SetVerbosity sets the logger verbosity
func (m *MultiLogger) SetVerbosity(v int) {
	m.verbosity = v
}

// SetDebugging sets the logger debugging verbosity
func (m *MultiLogger) SetDebugging(d int) {
	m.debug = d
}

// SetLevel sets the logger's overall level threshold
func (m *MultiLogger) SetLevel(l int) {
	m.level = l
}

// WriteEvent writes an event to all loggers on this MultiLogger
func (m *MultiLogger) WriteEvent(e *Event) {
	for _, l := range m.loggers {
		l.WriteEvent(e)
	}
}

// Shutdown shuts down all loggers on this MultiLogger
func (m *MultiLogger) Shutdown() {
	for _, l := range m.loggers {
		l.Shutdown()
	}
}

var _ Logger = &MultiLogger{}
