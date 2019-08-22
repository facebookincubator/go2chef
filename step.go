package go2chef

type Step interface {
	Component
	Download() error
	Execute() error
}

type StepLoader func(map[string]interface{}) (Step, error)

var (
	stepRegistry = make(map[string]StepLoader)
)

func RegisterStep(name string, s StepLoader) {
	if _, ok := stepRegistry[name]; ok {
		panic("step plugin " + name + " is already registered")
	}
	stepRegistry[name] = s
}

func GetStep(name string, config map[string]interface{}) (Step, error) {
	if s, ok := stepRegistry[name]; ok {
		return s(config)
	}
	return nil, &ErrComponentDoesNotExist{Component: name}
}
