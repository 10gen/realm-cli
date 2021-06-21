package terminal

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestOutputFormat(t *testing.T) {
	for _, tc := range []OutputFormat{
		// add all output formats here
		OutputFormatJSON,
		OutputFormatText,
	} {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidOutputFormat(tc), "must be valid output format")
		})
	}

	t.Run("Should have the correct type representation", func(t *testing.T) {
		assert.Equal(t, "String", OutputFormatText.Type())
	})

	t.Run("Should set its value correctly with a valid output format", func(t *testing.T) {
		tc := newOutputFormat()

		assert.Nil(t, tc.of.Set("json"))
		assert.Equal(t, "json", tc.of.String())

		assert.Nil(t, tc.of.Set(""))
		assert.Equal(t, "", tc.of.String())
	})

	t.Run("Should return an error when setting its value with an invalid output format", func(t *testing.T) {
		tc := newOutputFormat()
		assert.Equal(t, errors.New("unsupported value, use one of [json] instead"), tc.of.Set("eggcorn"))
	})
}

type outputFormatHolder struct {
	of *OutputFormat
}

func newOutputFormat() outputFormatHolder {
	var of OutputFormat
	return outputFormatHolder{&of}
}
