package flags

import (
	"errors"
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

type testType string

const (
	testType1 testType = "test-type-1"
	testType2 testType = "test-type-2"
)

const (
	value1         = "1-value"
	value2         = "2-value"
	valueWithComma = "3-value with, comma"
)

var (
	validValues         = []interface{}{value1, value2, valueWithComma}
	errUnsupportedValue = fmt.Errorf(`unsupported value, use one of ["1-value", "2-value", "3-value with, comma"] instead`)
)

func TestEnumSetValueSet(t *testing.T) {
	for _, tc := range []struct {
		description    string
		inputValue     string
		expectedValues []string
	}{
		{
			description:    "empty string should add no values",
			inputValue:     "",
			expectedValues: []string{},
		},
		{
			description:    "valid values should be added",
			inputValue:     fmt.Sprintf("%s,%s", value1, value2),
			expectedValues: []string{value1, value2},
		},
		{
			description:    "duplicates should only be added once",
			inputValue:     fmt.Sprintf("%s,%s", value1, value1),
			expectedValues: []string{value1},
		},
		{
			description:    "values with commas can be passed in with quotes",
			inputValue:     fmt.Sprintf(`%s,%s,"%s"`, value1, value2, valueWithComma),
			expectedValues: []string{value1, value2, valueWithComma},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			values := []string{}
			enumSetValue := NewEnumSet(&values, validValues)
			assert.Nil(t, enumSetValue.Set(tc.inputValue))
			assert.Equal(t, tc.expectedValues, values)
		})
	}
	t.Run("invalid values should cause an error", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, errUnsupportedValue, enumSetValue.Set("eggcorn"))
	})
	t.Run("custom type string values", func(t *testing.T) {
		for _, tc := range []struct {
			description    string
			inputValue     string
			expectedValues []string
		}{
			{
				description:    "empty string should add no values",
				inputValue:     "",
				expectedValues: []string{},
			},
			{
				description:    "valid values should be added",
				inputValue:     fmt.Sprintf("%v,%v", testType1, testType2),
				expectedValues: []string{string(testType1), string(testType2)},
			},
			{
				description:    "duplicates should only be added once",
				inputValue:     fmt.Sprintf("%v,%v", testType1, testType1),
				expectedValues: []string{string(testType1)},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				values := []string{}
				enumSetValue := NewEnumSet(&values, []interface{}{testType1, testType2})
				assert.Nil(t, enumSetValue.Set(tc.inputValue))
				assert.Equal(t, tc.expectedValues, values)
			})
		}
		t.Run("invalid values should cause an error", func(t *testing.T) {
			values := []string{}
			enumSetValue := NewEnumSet(&values, []interface{}{testType1, testType2})
			assert.Equal(t, errors.New(`unsupported value, use one of ["test-type-1", "test-type-2"] instead`), enumSetValue.Set("eggcorn"))
		})
	})
}

func TestEnumSetValueType(t *testing.T) {
	t.Run("should have type of string slice", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, "Set", enumSetValue.Type())
	})
}

func TestEnumSetValueString(t *testing.T) {
	for _, tc := range []struct {
		description    string
		values         []string
		expectedString string
	}{
		{
			description:    "no values should yield only brackets",
			values:         []string{},
			expectedString: "[]",
		},
		{
			description:    "values should be comma separated",
			values:         []string{value1, value2},
			expectedString: fmt.Sprintf("[%s,%s]", value1, value2),
		},
		{
			description:    "values with commas should be printed within quotes",
			values:         []string{valueWithComma},
			expectedString: fmt.Sprintf(`["%s"]`, valueWithComma),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var values []string
			enumSetValue := NewEnumSet(&values, validValues)
			assert.Nil(t, enumSetValue.set(tc.values...))
			assert.Equal(t, tc.expectedString, enumSetValue.String())
		})
	}
	t.Run("custom type string values", func(t *testing.T) {
		for _, tc := range []struct {
			description    string
			values         []string
			expectedString string
		}{
			{
				description:    "no values should yield only brackets",
				values:         []string{},
				expectedString: "[]",
			},
			{
				description:    "values should be comma separated",
				values:         []string{string(testType1), string(testType2)},
				expectedString: fmt.Sprintf(`[%v,%v]`, testType1, testType2),
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				var values []string
				enumSetValue := NewEnumSet(&values, []interface{}{testType1, testType2})
				assert.Nil(t, enumSetValue.set(tc.values...))
				assert.Equal(t, tc.expectedString, enumSetValue.String())
			})
		}
	})
}

func TestEnumSetValueAppend(t *testing.T) {
	for _, tc := range []struct {
		description    string
		initialValues  []string
		newValue       string
		expectedValues []string
	}{
		{
			description:    "a valid value should be added",
			initialValues:  []string{},
			newValue:       value1,
			expectedValues: []string{value1},
		},
		{
			description:    "a duplicate value should not be added",
			initialValues:  []string{value1},
			newValue:       value1,
			expectedValues: []string{value1},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			values := []string{}
			enumSetValue := NewEnumSet(&values, validValues)
			assert.Nil(t, enumSetValue.set(tc.initialValues...))
			assert.Nil(t, enumSetValue.Append(tc.newValue))
			assert.Equal(t, tc.expectedValues, values)
		})
	}
	t.Run("invalid values should cause an error", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, errUnsupportedValue, enumSetValue.Append("eggcorn"))
	})
	t.Run("custom type string values", func(t *testing.T) {
		for _, tc := range []struct {
			description    string
			initialValues  []string
			newValue       string
			expectedValues []string
		}{
			{
				description:    "a valid value should be added",
				initialValues:  []string{},
				newValue:       string(testType1),
				expectedValues: []string{string(testType1)},
			},
			{
				description:    "a duplicate value should not be added",
				initialValues:  []string{string(testType1)},
				newValue:       string(testType1),
				expectedValues: []string{string(testType1)},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				values := []string{}
				enumSetValue := NewEnumSet(&values, []interface{}{testType1, testType2})
				assert.Nil(t, enumSetValue.set(tc.initialValues...))
				assert.Nil(t, enumSetValue.Append(tc.newValue))
				assert.Equal(t, tc.expectedValues, values)
			})
		}
		t.Run("invalid values should cause an error", func(t *testing.T) {
			values := []string{}
			enumSetValue := NewEnumSet(&values, []interface{}{testType1, testType2})
			assert.Equal(t, errors.New(`unsupported value, use one of ["test-type-1", "test-type-2"] instead`), enumSetValue.Append("eggcorn"))
		})
	})
}

func TestEnumSetValueReplace(t *testing.T) {
	for _, tc := range []struct {
		description       string
		oldValues         []string
		newValues         []string
		expectedNewValues []string
	}{
		{
			description:       "values should be replaced if the new values are valid",
			oldValues:         []string{value1, value2},
			newValues:         []string{value2, valueWithComma},
			expectedNewValues: []string{value2, valueWithComma},
		},
		{
			description:       "duplicate values should not exist in new values",
			oldValues:         []string{value1, value2},
			newValues:         []string{value2, value2},
			expectedNewValues: []string{value2},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			values := []string{}
			enumSetValue := NewEnumSet(&values, validValues)
			assert.Nil(t, enumSetValue.set(tc.oldValues...))
			assert.Nil(t, enumSetValue.Replace(tc.newValues))
			assert.Equal(t, tc.expectedNewValues, values)
		})
	}
	t.Run("invalid values should cause an error", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, errUnsupportedValue, enumSetValue.Replace([]string{"eggcorn"}))
	})
	t.Run("custom type string values", func(t *testing.T) {
		for _, tc := range []struct {
			description       string
			oldValues         []string
			newValues         []string
			expectedNewValues []string
		}{
			{
				description:       "values should be replaced if the new values are valid",
				oldValues:         []string{string(testType1), string(testType2)},
				newValues:         []string{string(testType2)},
				expectedNewValues: []string{string(testType2)},
			},
			{
				description:       "duplicate values should not exist in new values",
				oldValues:         []string{string(testType1), string(testType2)},
				newValues:         []string{string(testType2), string(testType2)},
				expectedNewValues: []string{string(testType2)},
			},
		} {
			t.Run(tc.description, func(t *testing.T) {
				values := []string{}
				enumSetValue := NewEnumSet(&values, []interface{}{testType1, testType2})
				assert.Nil(t, enumSetValue.set(tc.oldValues...))
				assert.Nil(t, enumSetValue.Replace(tc.newValues))
				assert.Equal(t, tc.expectedNewValues, values)
			})
		}
		t.Run("invalid values should cause an error", func(t *testing.T) {
			values := []string{}
			enumSetValue := NewEnumSet(&values, []interface{}{testType1, testType2})
			assert.Equal(t, errors.New(`unsupported value, use one of ["test-type-1", "test-type-2"] instead`), enumSetValue.Replace([]string{"eggcorn"}))
		})
	})
}

func TestEnumSetValueGetSlice(t *testing.T) {
	t.Run("should return the values slice", func(t *testing.T) {
		values := []string{value1, value2}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, values, enumSetValue.GetSlice())
	})
}
