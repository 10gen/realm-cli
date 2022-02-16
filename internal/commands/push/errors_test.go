package push

import (
	"testing"

	"github.com/10gen/realm-cli/internal/cli/feedback"
	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestPushErrors(t *testing.T) {
	t.Run("err project not found should disable usage", func(t *testing.T) {
		var err = errProjectInvalid("", false)

		usageHider, ok := err.(feedback.ErrUsageHider)
		assert.True(t, ok, "expected project invalid error to hide usage")
		assert.True(t, usageHider.HideUsage(), "should hide usage")
	})
}
