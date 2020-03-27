package sanitycheck

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

// EnsureSuperuser checks that we're running as the administrator account.
// Shamelessly stolen from https://github.com/golang/go/issues/28804
func EnsureSuperuser(sc *SanityCheck) (FixFn, error) {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return func(*SanityCheck) error {
			return fmt.Errorf("SID Error: %s", err)
		}, err
	}
	token := windows.Token(0)

	isMember, err := token.IsMember(sid)
	if err != nil {
		return func(*SanityCheck) error {
			return fmt.Errorf("Token Membership Error: %s", err)
		}, err
	}

	if isMember {
		return nil, nil
	}
	return nil, errors.New("please elevate to admin")
}
