package go2chef

// Step defines the interface for go2chef execution steps
type Step interface {
	Component
	Download() error
	Execute() error
}

// StepLoader defines the function call interface for step loaders
type StepLoader func(map[string]interface{}) (Step, error)

var (
	stepRegistry = make(map[string]StepLoader)
)

// RegisterStep registers a new step plugin with go2chef
func RegisterStep(name string, s StepLoader) {
	if _, ok := stepRegistry[name]; ok {
		panic("step plugin " + name + " is already registered")
	}
	stepRegistry[name] = s
}

// GetStep gets a new step given a type and configuration
func GetStep(stepType string, config map[string]interface{}) (Step, error) {
	if s, ok := stepRegistry[stepType]; ok {
		return s(config)
	}
	return nil, &ErrComponentDoesNotExist{Component: stepType}
}
