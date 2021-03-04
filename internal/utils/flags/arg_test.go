package flags

import (
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestArg(t *testing.T) {
	t.Run("should print only name when value is nil", func(t *testing.T) {
		arg := Arg{Name: "test"}
		assert.Equal(t, " --test", arg.String())
	})

	t.Run("should print name and value when set", func(t *testing.T) {
		arg := Arg{"test", "value"}
		assert.Equal(t, " --test value", arg.String())
	})
}
