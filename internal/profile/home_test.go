package profile

import (
	"fmt"
	"testing"

	u "github.com/10gen/realm-cli/internal/utils/test"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestHomeDir(t *testing.T) {
	_, teardownHomeDir := u.SetupHomeDir("")
	defer teardownHomeDir()

	t.Run("Should return the home dir properly", func(t *testing.T) {
		home, err := homeDir()
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("./%s", profileDir), home)
	})
}
