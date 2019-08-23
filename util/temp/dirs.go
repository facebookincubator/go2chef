package temp

import (
	"fmt"
	"io/ioutil"
	"runtime"
)

var tmpDirs = make(map[string]string)

// TempDir creates a temporary directory registered for cleanup
func TempDir(dir, prefix string) (name string, err error) {
	name, err = ioutil.TempDir(dir, prefix)
	if err == nil {
		_, fn, ln, _ := runtime.Caller(1)
		tmpDirs[name] = fmt.Sprintf("%s:%d", fn, ln)
	}
	return
}
