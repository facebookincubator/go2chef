package go2chef

// Source defines the interface for source download components
type Source interface {
	Component
	DownloadToPath(path string) error
}

// SourceLoader represents factory functions for Sources
type SourceLoader func(map[string]interface{}) (Source, error)

var (
	sourceRegistry = make(map[string]SourceLoader)
)

// RegisterSource registers a new source plugin
func RegisterSource(name string, s SourceLoader) {
	if _, ok := sourceRegistry[name]; ok {
		panic("source plugin " + name + " is already registered")
	}
	sourceRegistry[name] = s
}

// GetSource gets the specified source plugin configured with the provided config map
func GetSource(name string, config map[string]interface{}) (Source, error) {
	if s, ok := sourceRegistry[name]; ok {
		return s(config)
	}
	return nil, &ErrComponentDoesNotExist{Component: name}
}
