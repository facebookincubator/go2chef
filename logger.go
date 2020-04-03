package go2chef

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"fmt"
	"strings"
)

type ExtraLoggingFields struct {
	StepName    string
	StepType    string
	StepCount   int
	ElapsedTime int
}

// Event provides a more structured way to log information from
// go2chef plugins.
type Event struct {
	Event       string
	Component   string
	Message     string
	ExtraFields *ExtraLoggingFields
}

// NewEvent returns a new event using the provided parameters
func NewEventWithExtraFields(event, component, message string, extrafields *ExtraLoggingFields) *Event {
	return &Event{
		Event:       event,
		Component:   component,
		Message:     message,
		ExtraFields: extrafields,
	}
}

// NewEvent returns a new event using the provided parameters
func NewEvent(event, component, message string) *Event {
	return &Event{
		Event:     event,
		Component: component,
		Message:   message,
	}
}

// Log level constants
const (
	LogLevelError = iota
	LogLevelInfo
	LogLevelDebug
)

// StringToLogLevel translates a string to a log level
func StringToLogLevel(s string) (int, error) {
	switch strings.ToLower(s) {
	case "debug":
		return LogLevelDebug, nil
	case "info":
		return LogLevelInfo, nil
	case "error":
		return LogLevelError, nil
	default:
		return LogLevelDebug, fmt.Errorf("log level %s does not exist", s)
	}
}

// LogLevelToString translates a log level value to a string
func LogLevelToString(l int) (string, error) {
	switch l {
	case LogLevelDebug:
		return "DEBUG", nil
	case LogLevelInfo:
		return "INFO", nil
	case LogLevelError:
		return "ERROR", nil
	default:
		return "", fmt.Errorf("log level %d is not valid", l)
	}
}

// Logger defines the interface for logging components.
type Logger interface {
	Component

	SetLevel(lvl int)
	SetDebug(dbg int)
	Debugf(dbg int, fmt string, args ...interface{})
	Infof(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})

	// WriteEvent writes an event object to this logger
	WriteEvent(e *Event)
	// Shutdown allows go2chef to wait for loggers to finish writes
	// if necessary (i.e. to remote endpoints)
	Shutdown()
}

// LoggerLoader defines the call signature for functions which
// return fully configured Logger instances
type LoggerLoader func(map[string]interface{}) (Logger, error)

var (
	logRegistry = make(map[string]LoggerLoader)
)

// RegisterLogger registers a new logging plugin
func RegisterLogger(name string, l LoggerLoader) {
	if _, ok := logRegistry[name]; ok {
		panic("log plugin " + name + " is already registered")
	}
	logRegistry[name] = l
}

// GetLogger gets a new instance of the Logger type specified by `name` and
// returns it configured as with config map[string]interface{}
func GetLogger(name string, config map[string]interface{}) (Logger, error) {
	if l, ok := logRegistry[name]; ok {
		return l(config)
	}
	return nil, &ErrComponentDoesNotExist{Component: name}
}

// GLOBAL LOGGER FUNCTIONALITY
//
// Provide a single central point-of-logging

var globalLogger Logger

// GetGlobalLogger gets an instance of the global logger
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		return NewMultiLogger([]Logger{})
	}
	return globalLogger
}

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(loggers []Logger) {
	globalLogger = NewMultiLogger(loggers)
}

// ShutdownGlobalLogger shuts down the global logger
func ShutdownGlobalLogger() {
	globalLogger.Shutdown()
}
