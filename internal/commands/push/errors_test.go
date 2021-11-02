package push

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestPushErrors(t *testing.T) {
	t.Run("err project not found should disable usage", func(t *testing.T) {
		var err error = errProjectInvalid{}

		_, ok := err.(cli.DisableUsage)
		assert.True(t, ok, "expected project invalid error to disable usage")
	})
}
