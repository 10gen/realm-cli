package flags

import (
	"testing"

	"github.com/spf13/pflag"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestOptionalStringFlag_Register(t *testing.T) {
	t.Run("should set user-defined value when flag is passed a value", func(t *testing.T) {
		flag := newOptStringFlag("test1", "defaultvalue")
		pflag.Var(&flag.Value, flag.Name , flag.Usage.String())
		pflag.Set(flag.Name, "test-string-value")
		pflag.Parse()

		f := pflag.Lookup(flag.Name)
		assert.NotNil(t, f)

		val, ok := f.Value.(*OptionalString)
		assert.Equal(t, true, ok)

		assert.Equal(t, true, val.IsSet)
		assert.Equal(t, "test-string-value", val.Value)
	})

	t.Run("should set default value when flag is not passed a value", func(t *testing.T) {
		flag := newOptStringFlag("test2", "defaultvalue")
		pflag.Var(&flag.Value, flag.Name , flag.Usage.String())
		pflag.Set(flag.Name, "")
		pflag.Parse()

		f := pflag.Lookup(flag.Name)
		assert.NotNil(t, f)

		val, ok := f.Value.(*OptionalString)
		assert.Equal(t, true, ok)

		assert.Equal(t, false, val.IsSet)
		assert.Equal(t, "defaultvalue", val.Value)
	})
}

func newOptStringFlag(name, defaultValue string) OptionalStringFlag {
	return OptionalStringFlag{
		Meta:  Meta{
			Name:      name,
			Usage:     Usage{
				Description:   "test template flag",
			},
		},
		Value: OptionalString{
			DefaultValue: defaultValue,
		},
	}
}
