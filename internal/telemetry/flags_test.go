package telemetry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestMode(t *testing.T) {
	for _, tc := range []Mode{
		// add all modes here
		ModeOn,
		ModeEmpty,
		ModeStdout,
		ModeOff,
	} {
		t.Run(fmt.Sprintf("%s should be valid", tc), func(t *testing.T) {
			assert.True(t, isValidMode(tc), "must be valid mode")
		})
	}

	t.Run("Should have the correct type representation", func(t *testing.T) {
		assert.Equal(t, "String", ModeOn.Type())
	})

	t.Run("Should set its value correctly with a valid output format", func(t *testing.T) {
		tc := newMode()

		assert.Nil(t, tc.m.Set("on"))
		assert.Equal(t, ModeOn, tc.m)

		assert.Nil(t, tc.m.Set(""))
		assert.Equal(t, ModeEmpty, tc.m)
	})

	t.Run("Should return an error when setting its value with an invalid output format", func(t *testing.T) {
		tc := newMode()
		assert.Equal(t, errors.New("unsupported value, use one of [on, off] instead"), tc.m.Set("eggcorn"))
	})
}

type modeHolder struct {
	m Mode
}

func newMode() modeHolder {
	var m Mode
	return modeHolder{m}
}
