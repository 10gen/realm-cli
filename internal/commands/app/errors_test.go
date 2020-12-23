package app

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestAppErrors(t *testing.T) {
	t.Run("err project exists should disable usage", func(t *testing.T) {
		var err error = errProjectExists{}

		_, ok := err.(cli.DisableUsage)
		assert.True(t, ok, "expected project exists error to disable usage")
	})
}
