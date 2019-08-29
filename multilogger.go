package go2chef

import (
	"math"
)

// MultiLogger is a fan-out logger for use as the central
// logging broker in go2chef
type MultiLogger struct {
	loggers []Logger
	debug   int
	level   int
}

// NewMultiLogger returns a MultiLogger with the provided list
// of loggers set up to receive logs.
func NewMultiLogger(loggers []Logger) *MultiLogger {
	return &MultiLogger{
		loggers: loggers,
		debug:   math.MaxInt64,
		level:   LogLevelDebug,
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

// Infof logs a formatted message at INFO level
func (m *MultiLogger) Infof(s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Infof(s, v...)
	}
}

// Debugf logs a formatted message at DEBUG level
func (m *MultiLogger) Debugf(dbg int, s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Debugf(dbg, s, v...)
	}
}

// SetLevel sets the logger's overall level threshold
func (m *MultiLogger) SetLevel(l int) {
	m.level = l
}

// SetDebug sets the logger's debug level threshold
func (m *MultiLogger) SetDebug(d int) {
	m.debug = d
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
