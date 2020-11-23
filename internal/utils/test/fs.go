package testutils

import (
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
)

// NewTempDir constructs a new temporary directory
// and returns the directory name along with a cleanup function
// or any error that occurred during the process
func NewTempDir(name string) (string, func(), error) {
	dir, err := ioutil.TempDir("", name)
	if err != nil {
		return "", nil, err
	}
	return dir, func() { os.RemoveAll(dir) }, nil
}

// SetupHomeDir sets up the $HOME directory for a test
// and returns the directory name along with a reset function
func SetupHomeDir(newHome string) (string, func()) {
	origHome := os.Getenv("HOME")
	if newHome == "" {
		newHome = "."
	}

	homedir.DisableCache = true
	_ = os.Setenv("HOME", newHome)

	return newHome, func() {
		homedir.DisableCache = false
		_ = os.Setenv("HOME", origHome)
	}
}
