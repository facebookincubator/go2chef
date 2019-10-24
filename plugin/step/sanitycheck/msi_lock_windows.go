package sanitycheck

import (
	"errors"
	"os/exec"
	"regexp"
	"runtime"
)

// IsMSIRunning is a simple check to see if the msiserver own lock which will
// prevent installation of other MSIs.
// It's possible for this service to run for a while after an MSI has terminated.
// However there is no point in trying when it won't even work, and this signal
// is as good as any.  ¯\_(ツ)_/¯
func IsMSIRunning(sc *SanityCheck) (FixFn, error) {
	if runtime.GOOS != "windows" {
		return nil, nil
	}

	return checkMSIRunning(sc)
}

func checkMSIRunning(*SanityCheck) (FixFn, error) {
	cmd := exec.Command("sc.exe", "query", "msiserver")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`\s+STATE\s+: \d+\s+([a-zA-Z]+)`)
	m := re.FindAllSubmatch(out, -1)
	state := string(m[0][1])
	if state != "STOPPED" {
		return nil, errors.New("msi installations are currently blocked")
	}

	return nil, nil
}

func init() {
	RegisterSanityCheck("msi_lock", IsMSIRunning)
}
