/*
Package plugconf implements a pluggable configuration store.

The way it works: install a plugin
*/
package plugconf

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

var (
	TagName            = "plugconf"
	ErrOutputNotStruct = errors.New("output is not a struct")
)

type ErrAlreadyRegistered struct {
	field string
	reg   *registration
}

func NewErrAlreadyRegistered(field string, reg *registration) *ErrAlreadyRegistered {
	return &ErrAlreadyRegistered{field: field, reg: reg}
}

func (e *ErrAlreadyRegistered) Error() string {
	return fmt.Sprintf("field %s is already registered from %s", e.field, e.reg.who)
}

type Processor func(field string, configData interface{}) error

type registration struct {
	who  string
	proc Processor
}

// PlugConf implements a pluggable configuration store
type PlugConf struct {
	tagName          string
	registeredFields map[string]registration
}

// NewPlugConf creates a new PlugConf with a given starting map
func NewPlugConf() *PlugConf {
	return &PlugConf{
		tagName:          TagName,
		registeredFields: make(map[string]registration),
	}
}

// Register puts a new output config in the
func (p *PlugConf) Register(field string, proc Processor) error {
	if reg, ok := p.registeredFields[field]; ok {
		return NewErrAlreadyRegistered(field, &reg)
	}
	_, fn, ln, _ := runtime.Caller(1)
	p.registeredFields[field] = registration{who: fn + ":" + strconv.Itoa(ln), proc: proc}
	return nil
}

// MustRegister is like Register but panics on failure
func (p *PlugConf) MustRegister(field string, proc Processor) {
	if err := p.Register(field, proc); err != nil {
		panic("error registering plugconf field: " + err.Error())
	}
}

// Process iterates through the registered configuration
// outputs and emits configurations for them
func (p *PlugConf) Process(config map[string]interface{}) error {
	for field, reg := range p.registeredFields {
		var toProc interface{}
		if fieldVal, ok := config[field]; ok {
			toProc = fieldVal
		}
		if err := reg.proc(field, toProc); err != nil {
			return err
		}
	}
	return nil
}
