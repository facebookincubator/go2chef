package group

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"fmt"
	"log"
	"sync"

	"github.com/facebookincubator/go2chef"
	"github.com/mitchellh/mapstructure"
)

// TypeName is the name of this step plugin
const TypeName = "go2chef.step.group"

// StepGroup defines a step that consists of other steps
//
// StepGroups download all resources in parallel and then execute
// steps sequentially. If you're doing a bunch of steps you
// probably want to use a `step_group` for it.
type StepGroup struct {
	GroupName string `mapstructure:"name"`
	logger    go2chef.Logger
	Steps     []go2chef.Step
}

func (g *StepGroup) String() string {
	return "<Step.Group:" + g.GroupName + ">"
}

// Name gets the name of this step instance
func (g *StepGroup) Name() string {
	return g.GroupName
}

// Type returns the type of this step instance
func (g *StepGroup) Type() string {
	return TypeName
}

// SetName sets the name of this step instance
func (g *StepGroup) SetName(n string) {
	g.GroupName = n
}

// Download runs the Download function of each substep in parallel
func (g *StepGroup) Download() (err error) {
	g.logger.WriteEvent(go2chef.NewEvent("STEP_GROUP_DOWNLOAD_START", TypeName, g.GroupName))
	defer func() {
		event := "STEP_GROUP_DOWNLOAD_COMPLETE"
		if err != nil {
			event = "STEP_GROUP_DOWNLOAD_FAILURE"
		}
		g.logger.WriteEvent(go2chef.NewEvent(event, TypeName, g.GroupName))
	}()
	var wg sync.WaitGroup
	errs := make(chan error, len(g.Steps))
	for _, s := range g.Steps {
		wg.Add(1)
		go func(st go2chef.Step, errs chan<- error) {
			defer wg.Done()
			if err := st.Download(); err != nil {
				errs <- err
			}
		}(s, errs)
	}
	wg.Wait()
	close(errs)

	count := 0
	for err := range errs {
		log.Printf("caught error: %s", err)
		count++
	}
	if count != 0 {
		return fmt.Errorf("errors during step group execution")
	}
	return nil
}

// Execute runs the Execute function of each substep in sequence
func (g *StepGroup) Execute() (err error) {
	g.logger.WriteEvent(go2chef.NewEvent("STEP_GROUP_EXECUTE_START", TypeName, g.GroupName))
	defer func() {
		event := "STEP_GROUP_EXECUTE_COMPLETE"
		if err != nil {
			event = "STEP_GROUP_EXECUTE_FAILURE"
		}
		g.logger.WriteEvent(go2chef.NewEvent(event, TypeName, g.GroupName))
	}()
	for _, s := range g.Steps {

		if err := s.Execute(); err != nil {
			return err
		}
	}
	return nil
}

// Loader provides an instantiation function for this step plugin
func Loader(config map[string]interface{}) (go2chef.Step, error) {
	// parse interior steps here
	structure := struct {
		Name  string                   `mapstructure:"name"`
		Steps []map[string]interface{} `mapstructure:"steps"`
	}{
		Steps: make([]map[string]interface{}, 0),
	}
	logger := go2chef.GetGlobalLogger()
	if err := mapstructure.Decode(config, &structure); err != nil {
		logger.Errorf("failed to parse configuration for %s: %s", TypeName, err)
		return nil, err
	}

	steps, err := go2chef.GetSteps(config)
	if err != nil {
		return nil, err
	}
	g := StepGroup{
		GroupName: structure.Name,
		logger:    go2chef.GetGlobalLogger(),
		Steps:     steps,
	}
	return &g, nil
}

var _ go2chef.Step = &StepGroup{}
var _ go2chef.StepLoader = Loader

func init() {
	go2chef.RegisterStep(TypeName, Loader)
}
