package testutils

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
)

// MustContainSubstring asserts the second provided arg is a substring of the first
// TODO(REALMC-7315): make this a go-cmp Option instead
func MustContainSubstring(t *testing.T, str, substr string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Errorf("expected %q to be a substring of %q, but it was not", substr, str)
	}
}

// MustNotBeNil asserts the provided arg is not nil
// TODO(REALMC-7315): make this a go-cmp Option instead
func MustNotBeNil(t *testing.T, o interface{}) {
	t.Helper()
	if o == nil {
		t.Error("expected value to not be <nil>, but it was")
	}
}

// MustMatch asserts the provided diff matches and fails the test if not
func MustMatch(t *testing.T, o string) {
	t.Helper()
	if o != "" {
		t.Error(o)
	}
}

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
