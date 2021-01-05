package go2chef

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"fmt"
	"runtime"
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
		debug:   int(^uint(0) >> 1),
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
		l.Errorf(stack2()+s, v...)
	}
}

// Infof logs a formatted message at INFO level
func (m *MultiLogger) Infof(s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Infof(stack2()+s, v...)
	}
}

// Debugf logs a formatted message at DEBUG level
func (m *MultiLogger) Debugf(dbg int, s string, v ...interface{}) {
	for _, l := range m.loggers {
		l.Debugf(dbg, stack2()+s, v...)
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

func stack2() string {
	_, f, l, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d::", f, l)
	}
	return "runtime-caller-err::"
}

var _ Logger = &MultiLogger{}
