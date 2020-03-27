// +build !windows

package sanitycheck

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"os/user"
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
