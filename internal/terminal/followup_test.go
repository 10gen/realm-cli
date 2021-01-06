package terminal

import (
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/google/go-cmp/cmp"
)

func TestNewFollowUpMessage( t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(followUpMessage{}), cmp.AllowUnexported(followUpMessage{}))

	for _, tc := range []struct {
		description string
		message string
		followUps []string
		expectedFollowUp followUpMessage
	} {
		{
			description: "Should return an empty follow up message"
				},
	}
}