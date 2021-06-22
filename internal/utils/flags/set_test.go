package flags

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestSet(t *testing.T) {
	t.Run("should have a set type", func(t *testing.T) {
		var values []string
		set := newStringSet(&values, nil)

		assert.Equal(t, "Set", set.Type())
	})

	t.Run("should initially print an empty array", func(t *testing.T) {
		var values []string
		set := newStringSet(&values, nil)

		assert.Equal(t, "[]", set.String())
	})

	t.Run("should print the set values as an array", func(t *testing.T) {
		values := []string{"a", "b"}
		set := newStringSet(&values, nil)

		assert.Equal(t, "[a,b]", set.String())
	})

	for _, tc := range []struct {
		description string
		val         string
		values      []string
	}{
		{
			description: "set no values from an empty string",
		},
		{
			description: "set values and omit any duplicates",
			val:         "one,two,one",
			values:      []string{"one", "two"},
		},
		{
			description: "set values written with quotes",
			val:         `one,two,"four,five"`,
			values:      []string{"four,five", "one", "two"},
		},
	} {
		t.Run("with no valid values should "+tc.description, func(t *testing.T) {
			var values []string
			set := newStringSet(&values, nil)

			assert.Nil(t, set.Set(tc.val))
			assert.Equal(t, tc.values, values)
		})
	}

	for _, tc := range []struct {
		description string
		val         string
		err         error
		values      []string
	}{
		{
			description: "set no values from an empty string",
		},
		{
			description: "set values and omit any duplicates",
			val:         "one,two,one,three",
			values:      []string{"one", "three", "two"},
		},
		{
			description: "set values written with quotes",
			val:         `one,two,three,"four,five"`,
			values:      []string{"four,five", "one", "three", "two"},
		},
		{
			description: "error if any non valid value is set",
			val:         "one,two,three,four",
			err:         errors.New("'four' is an unsupported value, try instead one of ['one', 'two', 'three', 'four,five']"),
		},
	} {
		t.Run("with a set of valid values should "+tc.description, func(t *testing.T) {
			var values []string
			set := newStringSet(&values, []string{"one", "two", "three", "four,five"})

			assert.Equal(t, tc.err, set.Set(tc.val))
			assert.Equal(t, tc.values, values)
		})
	}

	for _, tc := range []struct {
		description string
		val         string
		values      []string
	}{
		{
			description: "append no values from an empty string",
			values:      []string{"one", "two"},
		},
		{
			description: "append no values if duplicates are passed in",
			val:         "two",
			values:      []string{"one", "two"},
		},
		{
			description: "append values",
			val:         "three",
			values:      []string{"one", "three", "two"},
		},
	} {
		t.Run("with no valid values should "+tc.description, func(t *testing.T) {
			var values []string
			set := newStringSet(&values, nil)

			assert.Nil(t, set.Set("one,two")) // add some data

			assert.Nil(t, set.Append(tc.val))
			assert.Equal(t, tc.values, values)
		})
	}

	for _, tc := range []struct {
		description string
		val         string
		err         error
		values      []string
	}{
		{
			description: "append no values from an empty string",
		},
		{
			description: "append no values if duplicates are passed in",
			val:         "two",
			values:      []string{"one", "two"},
		},
		{
			description: "append values",
			val:         "three",
			values:      []string{"one", "three", "two"},
		},
		{
			description: "error if invalid value is appended",
			val:         "eggcorn",
			err:         errors.New("'four' is an unsupported value, try instead one of ['one', 'two', 'three']"),
		},
	} {
		t.Run("with a set of valid values should "+tc.description, func(t *testing.T) {
			var values []string
			set := newStringSet(&values, []string{"one", "two", "three"})

			assert.Nil(t, set.Set("one,two")) // add some data
		})
	}

	for _, tc := range []struct {
		description string
		vals        []string
		values      []string
	}{
		{
			description: "replace with an empty array",
			values:      []string{},
		},
		{
			description: "replace values",
			vals:        []string{"three", "four"},
			values:      []string{"four", "three"},
		},
	} {
		t.Run("with no valid values should "+tc.description, func(t *testing.T) {
			var values []string
			set := newStringSet(&values, nil)

			assert.Nil(t, set.Set("one,two")) // add some data

			assert.Nil(t, set.Replace(tc.vals))
			assert.Equal(t, tc.values, values)
		})
	}

	for _, tc := range []struct {
		description string
		vals        []string
		err         error
		values      []string
	}{
		{
			description: "replace with an empty array",
			values:      []string{},
		},
		{
			description: "replace values",
			vals:        []string{"one", "three"},
			values:      []string{"one", "three"},
		},
		{
			description: "error if invalid value is included in replace set",
			vals:        []string{"one", "three", "four"},
			err:         errors.New("'four' is an unsupported value, try instead one of ['one', 'two', 'three']"),
		},
	} {
		t.Run("with a set of valid values should "+tc.description, func(t *testing.T) {
			var values []string
			set := newStringSet(&values, []string{"one", "two", "three"})

			assert.Nil(t, set.Set("one,two")) // add some data

			assert.Equal(t, tc.err, set.Replace(tc.vals))
			assert.Equal(t, tc.values, values)
		})
	}
}
