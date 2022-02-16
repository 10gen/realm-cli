package app

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/feedback"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestAppErrors(t *testing.T) {
	t.Run("err project exists should hide usage", func(t *testing.T) {
		var err error = errProjectExists("")

		usageHider, ok := err.(feedback.ErrUsageHider)
		assert.True(t, ok, "expected project exists error to hide usage")
		assert.True(t, usageHider.HideUsage(), "should hide usage")
	})
}
