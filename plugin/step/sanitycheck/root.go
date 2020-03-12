package sanitycheck

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"errors"
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
