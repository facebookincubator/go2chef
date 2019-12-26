// +build windows

package sanitycheck

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"errors"
)

// EnsureSuperuser checks that we're running as the administrator account.
// Shamelessly stolen from https://github.com/golang/go/issues/28804
func EnsureSuperuser(sc *SanityCheck) (FixFn, error) {
	return nil, errors.New("use go2chef.winsanitycheck for windows go2chef configs")
}
