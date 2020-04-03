package cli

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"strconv"
	"time"

	"github.com/facebookincubator/go2chef/util/temp"

	"github.com/facebookincubator/go2chef"
	"github.com/facebookincubator/go2chef/plugin/logger/stdlib"
	"github.com/spf13/pflag"
)

var (
	// DefaultConfigSource sets the go2chef CLI default configuration source type
	DefaultConfigSource = "go2chef.config_source.local"
	// DefaultLogLevel sets the go2chef CLI default logging level
	DefaultLogLevel = go2chef.LogLevelDebug
	logger          go2chef.Logger
)

func init() {
}

// Go2ChefCLI is the CLI entry point for go2chef
type Go2ChefCLI struct {
	flags            *pflag.FlagSet
	configSourceName string
	logLevel         string
	logDebugLevel    int
	preserveTemp     bool
}

// Option defines the interface for CLI option functions
type Option func(cli *Go2ChefCLI)

// WithFlagSet is an Option to set a custom FlagSet
func WithFlagSet(set *pflag.FlagSet) Option {
	return func(cli *Go2ChefCLI) {
		cli.flags = set
	}
}

// NewGo2ChefCLI configures a Go2ChefCLI instance
func NewGo2ChefCLI(opts ...Option) *Go2ChefCLI {
	cli := &Go2ChefCLI{
		flags: pflag.NewFlagSet("go2chef", pflag.ExitOnError),
	}
	for _, opt := range opts {
		opt(cli)
	}

	logLevel, err := go2chef.LogLevelToString(DefaultLogLevel)
	if err != nil {
		panic("invalid go2chef.cli.DefaultLogLevel compiled in")
	}
	cli.flags.StringVarP(&cli.configSourceName, "config-source", "C", DefaultConfigSource, "name of the configuration source to use")
	cli.flags.StringVarP(&cli.logLevel, "log-level", "l", logLevel, "log level")
	cli.flags.BoolVar(&cli.preserveTemp, "preserve-temp", false, "preserve temporary directories from this run")
	return cli
}

// Run kicks off the execution of go2chef
func (g *Go2ChefCLI) Run(argv []string) int {
	// Set early config flags and parse. As we build our
	// own pflag.FlagSet plugins using pflag.*Var() functions
	// won't be able to pollute this.
	go2chef.InitializeConfigSourceFlags(g.flags)
	if err := g.flags.Parse(argv); err != nil {
		return 1
	}

	// Pull in early log level config from flags
	logLevel, err := go2chef.StringToLogLevel(g.logLevel)
	if err != nil {
		go2chef.EarlyLogger.Printf("--log-level/-l value `%s` is invalid: %s", g.logLevel, err)
		return 1
	}

	// Add stdlib early logger
	early := stdlib.NewFromLogger(go2chef.EarlyLogger, logLevel, g.logDebugLevel)

	// Load actual configuration
	cfg, err := go2chef.GetConfig(g.configSourceName, early)
	if err != nil {
		early.Errorf("config error: %s", err)
		return 1
	}

	// Wire up central logging
	go2chef.InitGlobalLogger(cfg.Loggers)
	logger = go2chef.GetGlobalLogger()

	logger.WriteEvent(&go2chef.Event{
		Event:     "LOGGING_INITIALIZED",
		Component: "go2chef.cli",
	})

	defer temp.Cleanup(g.preserveTemp)

	all_start := time.Now()
	for i, step := range cfg.Steps {
		start := time.Now()
		eventStartStep(i, step.Name(), step.Type())
		if err := step.Download(); err != nil {
			eventFailStep(i, err, step.Name(), step.Type())
			return 1
		}
		if err := step.Execute(); err != nil {
			eventFailStep(i, err, step.Name(), step.Type())
			return 1
		}
		elapsed := int(time.Since(start).Seconds())
		eventFinishStep(i, elapsed, step.Name(), step.Type())
	}

	all_elapsed := int(time.Since(all_start).Seconds())
	eventFinishAllSteps(len(cfg.Steps), all_elapsed)
	go2chef.ShutdownGlobalLogger()
	return 0
}

func eventStartStep(idx int, step_name, step_type string) {
	logger.WriteEvent(&go2chef.Event{
		Event:     "STEP_" + strconv.Itoa(idx) + "_START " + step_type + ":" + "'" + step_name + "'",
		Component: "go2chef.cli",
	})
}

func eventFailStep(idx int, err error, step_name, step_type string) {
	logger.WriteEvent(&go2chef.Event{
		Event:     "STEP_" + strconv.Itoa(idx) + "_FAILURE " + step_type + ":" + "'" + step_name + "'",
		Component: "go2chef.cli",
		Message:   err.Error(),
	})
}

func eventFinishStep(idx, elapsed int, step_name, step_type string) {
	logger.WriteEvent(&go2chef.Event{
		Event:     "STEP_" + strconv.Itoa(idx) + "_COMPLETE " + step_type + ":" + "'" + step_name + "'",
		Component: "go2chef.cli",
		Message:   "completed successfully in " + strconv.Itoa(elapsed) + " second(s)",
		ExtraFields: (&go2chef.ExtraLoggingFields{
			StepName:    step_name,
			StepType:    step_type,
			ElapsedTime: elapsed,
		}),
	})
}

func eventFinishAllSteps(steps int, elapsed int) {
	logger.WriteEvent(&go2chef.Event{
		Event:     "ALL_STEPS_COMPLETE",
		Component: "go2chef.cli",
		Message:   strconv.Itoa(steps) + " step(s) completed successfully in " + strconv.Itoa(elapsed) + " second(s)",
		ExtraFields: (&go2chef.ExtraLoggingFields{
			StepName:    "All_Steps",
			StepCount:   steps,
			ElapsedTime: elapsed,
		}),
	})
}
