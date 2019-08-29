package sanitycheck

import (
	"errors"
	"fmt"
	"os/user"
	"runtime"
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
	switch runtime.GOOS {
	case "windows":
		return nil, fmt.Errorf("not yet implemented")
	default:
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		if u.Username == UnixSuperuserUsername && u.Uid == UnixSuperuserUID {
			return nil, nil
		}
		return nil, ErrNotSuperuser
	}
}
