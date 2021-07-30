package flags

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestOptionalStringSet(t *testing.T) {
	defaultValue := "default"
	opt := OptionalString{
		DefaultValue: defaultValue,
	}
	for _, tc := range []struct {
		description string
		input       string
		expectedIsSet bool
		expectedValue string
	}{
		{"no input", "", false, defaultValue},
		{"string input", "hello-world", true, "hello-world"},
		{"string number input", "12", true, "12"},
	} {
		t.Run("should parse "+tc.description, func(t *testing.T) {
			assert.Nil(t, opt.Set(tc.input))

			assert.Equal(t, tc.expectedIsSet, opt.IsSet)
			assert.Equal(t, tc.expectedValue, opt.Value)
		})
	}
}
