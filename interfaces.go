package go2chef

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import "fmt"

// Component defines the interface for go2chef components (plugins)
type Component interface {
	fmt.Stringer
	SetName(string)
	Name() string
	Type() string
}
