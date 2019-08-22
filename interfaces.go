package go2chef

import "fmt"

// Component defines the interface for go2chef components (plugins)
type Component interface {
	fmt.Stringer
	SetName(string)
	Name() string
	Type() string
}
