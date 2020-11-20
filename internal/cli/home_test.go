package cli

import (
	"fmt"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"

	"github.com/google/go-cmp/cmp"
)

func TestHomeDir(t *testing.T) {
	_, teardownHomeDir := u.SetupHomeDir("")
	defer teardownHomeDir()

	t.Run("Should return the home dir properly", func(t *testing.T) {
		home, err := homeDir()
		u.MustMatch(t, cmp.Diff(nil, err))
		u.MustMatch(t, cmp.Diff(fmt.Sprintf("./%s", profileDir), home))
	})
}
