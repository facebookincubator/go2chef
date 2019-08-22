package go2chef

import (
	"github.com/oko/logif"
)

// Event provides a more structured way to log information from
// go2chef plugins.
type Event struct {
	Event     string
	Component string
	Message   string
}

// NewEvent returns a new event using the provided parameters
func NewEvent(event, component, message string) *Event {
	return &Event{
		Event:     event,
		Component: component,
		Message:   message,
	}
}

// Logger defines the interface for logging components.
type Logger interface {
	Component
	logif.Logger
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
