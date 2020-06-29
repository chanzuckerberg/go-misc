package osutil

import (
	"io/ioutil"
	"os"
	"regexp"
	"runtime"

	"github.com/pkg/errors"
)

var reWSL = regexp.MustCompile("microsoft|Microsoft|WSL")

func IsWSL() (bool, error) {
	if runtime.GOOS != "linux" {
		return false, nil
	}

	contents, err := ioutil.ReadFile("/proc/sys/kernel/osrelease")
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, errors.Wrap(err, "could not detect WSL")
	}

	return reWSL.Match(contents), nil
}
