package plugconf

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

var (
	TagName = "plugconf"
	ErrOutputNotStruct = errors.New("output is not a struct")
)

type PlugConf struct {
	tagName string
	remaining map[string]interface{}
	registry []interface{}
	registeredFields map[string]string
}

func NewPlugConf(start map[string]interface{}) *PlugConf {
	return &PlugConf{
		tagName: TagName,
		remaining:start,
		registry: nil,
		registeredFields: make(map[string]string),
	}
}

func (p *PlugConf) Register(output interface{}) error {
	t := reflect.TypeOf(output)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ErrOutputNotStruct
	}
	type tagReg struct {
		tag string
		name string
	}
	var toReg []tagReg
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get(p.tagName)
		if tag != "" {
			if ex, ok := p.registeredFields[tag]; ok {
				return fmt.Errorf("field %s is already registered for type %s", tag, ex)
			} else {
				toReg = append(toReg, tagReg{tag:tag, name: t.Name()})
			}
		}
	}
	p.registry = append(p.registry, output)
	for _, tr := range toReg{
		p.registeredFields[tr.tag] = tr.name
	}
	return nil
}

func (p *PlugConf) Process() error {
	for _, output := range p.registry {
		if err := p.process(output); err != nil {
			return err
		}
	}
	return nil
}

func (p *PlugConf) process(output interface{}) error {
	var md mapstructure.Metadata
	mcfg := &mapstructure.DecoderConfig{
		Metadata:         &md,
		Result:           output,
		TagName:          TagName,
	}
	dec, err := mapstructure.NewDecoder(mcfg)
	if err != nil {
		return err
	}
	if err := dec.Decode(p.remaining); err != nil {
		return err
	}
	for _, delKey := range md.Keys {
		delete(p.remaining, delKey)
	}
	return nil
}