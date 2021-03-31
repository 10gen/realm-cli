package flags

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestDateSet(t *testing.T) {
	for _, tc := range []struct {
		description string
		input       string
		output      string
	}{
		{"a timestamp at millis with time zone", "2021-06-22T07:54:42.123-0500", "2021-06-22T07:54:42.123-0500"},
		{"a timestamp at millis", "2021-06-22T07:54:42.123", "2021-06-22T07:54:42.123+0000"},
		{"a timestamp with time zone", "2021-06-22T07:54:42-0500", "2021-06-22T07:54:42.000-0500"},
		{"a timestamp", "2021-06-22T07:54:42", "2021-06-22T07:54:42.000+0000"},
		{"a timestamp at minutes with zone", "2021-06-22T07:54-0500", "2021-06-22T07:54:00.000-0500"},
		{"a timestamp at minutes", "2021-06-22T07:54", "2021-06-22T07:54:00.000+0000"},
		{"a timestamp at hours with time zone", "2021-06-22T07-0500", "2021-06-22T07:00:00.000-0500"},
		{"a timestamp at hours", "2021-06-22T07", "2021-06-22T07:00:00.000+0000"},
		{"a timestamp at days with time zone", "2021-06-22-0500", "2021-06-22T00:00:00.000-0500"},
		{"a timestamp at days", "2021-06-22", "2021-06-22T00:00:00.000+0000"},
	} {
		t.Run("should parse "+tc.description, func(t *testing.T) {
			date := new(Date)

			assert.Nil(t, date.Set(tc.input))

			assert.Equal(t, tc.output, date.String())
		})
	}
}
