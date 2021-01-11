package flags

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

const (
	value1         = "1-value"
	value2         = "2-value"
	valueWithComma = "3-value with, comma"
)

var (
	validValues         = []string{value1, value2, valueWithComma}
	errUnsupportedValue = fmt.Errorf(`unsupported value, use one of ["1-value", "2-value", "3-value with, comma"] instead`)
)

func TestEnumSetValueSet(t *testing.T) {
	for _, tc := range []struct {
		description    string
		inputValue     string
		expectedValues []string
	}{
		{
			description:    "Empty string should add no values",
			inputValue:     "",
			expectedValues: []string{},
		},
		{
			description:    "Valid values should be added",
			inputValue:     fmt.Sprintf("%s,%s", value1, value2),
			expectedValues: []string{value1, value2},
		},
		{
			description:    "Duplicates should only be added once",
			inputValue:     fmt.Sprintf("%s,%s", value1, value1),
			expectedValues: []string{value1},
		},
		{
			description:    "Values with commas can be passed in with quotes",
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
	t.Run("Invalid values should cause an error", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, errUnsupportedValue, enumSetValue.Set("eggcorn"))
	})
}

func TestEnumSetValueType(t *testing.T) {
	t.Run("EnumSetValue should have type of stringSlice", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, "enumSet", enumSetValue.Type())
	})
}

func TestEnumSetValueString(t *testing.T) {
	for _, tc := range []struct {
		description    string
		values         []string
		expectedString string
	}{
		{
			description:    "No values should yield only brackets",
			values:         []string{},
			expectedString: "[]",
		},
		{
			description:    "Values should be comma separated",
			values:         []string{value1, value2},
			expectedString: fmt.Sprintf("[%s,%s]", value1, value2),
		},
		{
			description:    "Values with commas should be printed within quotes",
			values:         []string{valueWithComma},
			expectedString: fmt.Sprintf(`["%s"]`, valueWithComma),
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			var values []string
			enumSetValue := NewEnumSet(&values, tc.values)
			assert.Nil(t, enumSetValue.set(tc.values...))
			assert.Equal(t, tc.expectedString, enumSetValue.String())
		})
	}
}

func TestEnumSetValueAppend(t *testing.T) {
	for _, tc := range []struct {
		description    string
		initialValues  []string
		newValue       string
		expectedValues []string
	}{
		{
			description:    "A Valid value should be added",
			initialValues:  []string{},
			newValue:       value1,
			expectedValues: []string{value1},
		},
		{
			description:    "A duplicate value should not be added",
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
	t.Run("Invalid values should cause an error", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, errUnsupportedValue, enumSetValue.Append("eggcorn"))
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
			description:       "Values should be replaced if the new values are valid",
			oldValues:         []string{value1, value2},
			newValues:         []string{value2, valueWithComma},
			expectedNewValues: []string{value2, valueWithComma},
		},
		{
			description:       "Duplicate values should not exist in new values",
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
	t.Run("Invalid values should cause an error", func(t *testing.T) {
		values := []string{}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, errUnsupportedValue, enumSetValue.Replace([]string{"eggcorn"}))
	})
}

func TestEnumSetValueGetSlice(t *testing.T) {
	t.Run("EnumSetValue should return the values slice", func(t *testing.T) {
		values := []string{value1, value2}
		enumSetValue := NewEnumSet(&values, validValues)
		assert.Equal(t, values, enumSetValue.GetSlice())
	})
}
