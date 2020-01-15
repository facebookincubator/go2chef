// +build !windows

package winsanitycheck

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"errors"
	"os/user"
)

// ErrNotSuperuser is the error raised when not running as a superuser
var (
	ErrNotSuperuser = errors.New("not running as superuser")
)

// Unix superuser variables to allow testing
var (
	UnixSuperuserUsername = "root"
	UnixSuperuserUID      = "0"
)

// EnsureSuperuser checks that we're running as superuser
func EnsureSuperuser(sc *SanityCheck) (FixFn, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	if u.Username == UnixSuperuserUsername && u.Uid == UnixSuperuserUID {
		return nil, nil
	}
	return nil, ErrNotSuperuser
}
