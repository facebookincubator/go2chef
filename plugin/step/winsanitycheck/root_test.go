package winsanitycheck

/*
	Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
*/

import (
	"os/user"
	"testing"
)

func TestEnsureSuperuser(t *testing.T) {
	curUser, err := user.Current()
	if err != nil {
		t.Fatal("failed to call user.Current() for test setup")
	}
	UnixSuperuserUID = curUser.Uid
	UnixSuperuserUsername = curUser.Username

	fix, err := EnsureSuperuser(nil)
	if err != nil {
		t.Fatalf("got an error from EnsureSuperuser: %s", err)
	}
	if fix != nil {
		t.Fatalf("got a fix function from EnsureSuperuser when we shouldn't have")
	}

	UnixSuperuserUsername = "thisshouldnotexist"

	fix, err = EnsureSuperuser(nil)
	if err != nil && err != ErrNotSuperuser {
		t.Fatalf("error from EnsureSuperuser but not expected ErrNotSuperuser")
	}
	if err == nil {
		t.Fatalf("no error from EnsureSuperuser but expected ErrNotSuperuser")
	}
	if fix != nil {
		t.Fatalf("got a fix function from EnsureSuperuser when we shouldn't have")
	}
}

func TestEnsureSuperuserWhenNotSuperuser(t *testing.T) {
	UnixSuperuserUsername = "thisshouldnotexist"

	fix, err := EnsureSuperuser(nil)
	if err != nil && err != ErrNotSuperuser {
		t.Fatalf("error from EnsureSuperuser but not expected ErrNotSuperuser")
	}
	if err == nil {
		t.Fatalf("no error from EnsureSuperuser but expected ErrNotSuperuser")
	}
	if fix != nil {
		t.Fatalf("got a fix function from EnsureSuperuser when we shouldn't have")
	}
}
