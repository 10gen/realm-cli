package terminal

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
)

func TestNewTable(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(table{}), cmp.AllowUnexported(table{}))

	for _, tc := range []struct {
		description   string
		message       string
		header        []string
		data          []map[string]interface{}
		expectedTable table
	}{
		{
			description: "Should return empty table if there is nil for the header",
			header:      nil,
		},
		{
			description: "Should return empty table if there is no header information",
			header:      []string{},
		},
		{
			description: "Should return a table with only headers if there is one and no data",
			message:     "a table message",
			header:      []string{"should", "show", "up", "only"},
			expectedTable: table{
				"a table message",
				[]string{"should", "show", "up", "only"},
				make([]map[string]string, 0),
				map[string]int{
					"should": 6,
					"show":   4,
					"up":     2,
					"only":   4,
				},
			},
		},
		{
			description: "Should return a table if there is both a header and data",
			message:     "a table message",
			header:      []string{"test"},
			data: []map[string]interface{}{
				{
					"test":   "should show up",
					"int":    1,
					"float":  1.000,
					"float2": 1.234567890123,
					"array":  []string{"1", "2", "3"},
					"nil":    nil,
				},
			},
			expectedTable: table{
				"a table message",
				[]string{"test"},
				[]map[string]string{{
					"test": "should show up",
				}},
				map[string]int{
					"test": 14,
				},
			},
		},
	} {
		t.Run(tc.description, func(*testing.T) {
			table := newTable(tc.message, tc.header, tc.data)
			assert.Equal(t, tc.expectedTable, table)
		})
	}
}

func TestTableMessage(t *testing.T) {
	t.Run("Should return an empty string and an error if the table has no headers", func(t *testing.T) {
		table := newTable("", nil, nil)
		message, err := table.Message()
		assert.Equal(t, message, "")
		assert.Equal(t, err.Error(), "cannot create a table without headers")
	})

	for _, tc := range []struct {
		description     string
		message         string
		header          []string
		data            []map[string]interface{}
		expectedMessage string
	}{
		{
			description: "Should print only a header for a table with no data",
			message:     "a table message",
			header:      []string{"header", "only", "no", "data"},
			expectedMessage: fmt.Sprintf(`a table message
  %s  %s  %s  %s
  ------  ----  --  ----
`,
				[]interface{}{
					color.New(color.Bold).SprintFunc()("header"),
					color.New(color.Bold).SprintFunc()("only"),
					color.New(color.Bold).SprintFunc()("no"),
					color.New(color.Bold).SprintFunc()("data"),
				}...,
			),
		},

		{
			description: "Should return correctly formatted values in the table, not create new columns, and not print empty rows",
			message:     "a table message",
			header:      []string{"arrays", "floats", "ints", "maps/objects", "strings", "sparse", "errors"},
			data: []map[string]interface{}{
				{
					"ints":    1,
					"floats":  12.34,
					"strings": "tester string",
					"arrays":  []string{"1", "test", "this"},
					"maps/objects": struct {
						test int
						this float64
					}{test: 1234, this: 12.345},
				},
				{},
				{
					"ints":    000000001,
					"floats":  12.34,
					"strings": "test",
					"arrays":  []int{1, 2, 3, 4},
					"maps/objects": map[string]int{
						"test": 1,
						"this": 2,
					},
					"sparse": "hello",
					"errors": errors.New("new error"),
				},
				{
					"ints":    -0002,
					"floats":  0000123.340000,
					"extra":   "shouldn't show up",
					"strings": "hello",
					"arrays":  []float64{1, 2.3, 4.5555555555, 6},
					"maps/objects": map[string][]int{
						"what happens": {1, 2, 3},
					},
				},
				{},
			},
			expectedMessage: fmt.Sprintf(
				strings.Join([]string{
					"a table message",
					"  %s                  %s  %s  %s               %s        %s  %s   ",
					"  ----------------------  ------  ----  -------------------------  -------------  ------  ---------",
					"  [1 test this]           12.34   1     {test:1234 this:12.345}    tester string                   ",
					"  [1 2 3 4]               12.34   1     map[test:1 this:2]         test           hello   new error",
					"  [1 2.3 4.5555555555 6]  123.34  -2    map[what happens:[1 2 3]]  hello                           ",
				}, "\n"),
				[]interface{}{
					color.New(color.Bold).SprintFunc()("arrays"),
					color.New(color.Bold).SprintFunc()("floats"),
					color.New(color.Bold).SprintFunc()("ints"),
					color.New(color.Bold).SprintFunc()("maps/objects"),
					color.New(color.Bold).SprintFunc()("strings"),
					color.New(color.Bold).SprintFunc()("sparse"),
					color.New(color.Bold).SprintFunc()("errors"),
				}...,
			),
		},
	} {
		t.Run(tc.description, func(*testing.T) {
			table := newTable(tc.message, tc.header, tc.data)
			message, err := table.Message()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedMessage, message)
		})
	}
}

func TestTableMessageNoBold(t *testing.T) {
	for _, tc := range []struct {
		description     string
		message         string
		header          []string
		data            []map[string]interface{}
		expectedMessage string
	}{
		{
			description: "Should not bold the headers if color is off",
			message:     "a table message",
			header:      []string{"should", "not", "be", "bold"},
			expectedMessage: `a table message
  should  not  be  bold
  ------  ---  --  ----
`,
		},
		{
			description: "Should not bold the headers if color is off even if there is data",
			header:      []string{"should", "not", "be", "bold"},
			message:     "a table message",
			data: []map[string]interface{}{
				{
					"should": 123,
					"be":     "not bold!",
				},
			},
			expectedMessage: `a table message
  should  not  be         bold
  ------  ---  ---------  ----
  123          not bold!      `,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			color.NoColor = true
			defer func() { color.NoColor = false }()

			table := newTable(tc.message, tc.header, tc.data)
			message, err := table.Message()
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedMessage, message)
		})
	}
}

func TestTablePayload(t *testing.T) {
	t.Run("Payload should return an error with a table without a header", func(t *testing.T) {
		table := newTable("", nil, nil)
		payloadKeys, payloadData, err := table.Payload()
		assert.Nil(t, payloadKeys)
		assert.Nil(t, payloadData)
		assert.Equal(t, "cannot create a table without headers", err.Error())
	})

	t.Run("Payload should work with a valid table", func(t *testing.T) {
		message := "a table message"
		headers := []string{"test", "this", "data"}

		table := newTable(message, headers, []map[string]interface{}{{
			"test": 123,
			"this": "456",
			"data": []string{"7", "8 9", "10!"},
		}})

		payloadKeys, payloadData, err := table.Payload()
		assert.Nil(t, err)
		assert.Equal(t, tableFields, payloadKeys)

		data := []map[string]string{
			{
				"test": "123",
				"this": "456",
				"data": "[7 8 9 10!]",
			},
		}

		assert.Equal(t, message, payloadData[logFieldMessage])
		assert.Equal(t, headers, payloadData[logFieldHeaders])
		assert.Equal(t, data, payloadData[logFieldData])
	})
}

func TestParseValue(t *testing.T) {
	for _, tc := range []struct {
		description    string
		value          interface{}
		expectedString string
	}{
		{
			description:    "a nil value as an empty string",
			value:          nil,
			expectedString: "",
		},
		{
			description:    "an empty string as an empty string",
			value:          "",
			expectedString: "",
		},
		{
			description:    "the string 'strings' as 'strings'",
			value:          "string",
			expectedString: "string",
		},
		{
			description:    "an empty generic slice as '[]'",
			value:          []interface{}{},
			expectedString: "[]",
		},
		{
			description:    "an empty string slice as '[]'",
			value:          []string{},
			expectedString: "[]",
		},
		{
			description:    "a slice of strings as a non-comma-separated list of strings",
			value:          []string{"slice", "of", "strings"},
			expectedString: "[slice of strings]",
		},
		{
			description:    "a slice of ints as a non-comma-separated list of ints",
			value:          []int{1, 2, 3},
			expectedString: "[1 2 3]",
		},
		{
			description:    "a generic slice as a non-comma-separated list with each value properly parsed",
			value:          []interface{}{1, "2", []int{3, 3, 3}},
			expectedString: "[1 2 [3 3 3]]",
		},
		{
			description:    "the integer 42 as '42'",
			value:          42,
			expectedString: "42",
		},
		{
			description:    "the negative integer -42 as '-42'",
			value:          -42,
			expectedString: "-42",
		},
		{
			description:    "the whole number float 42.0 as '42'",
			value:          42.0,
			expectedString: "42",
		},
		{
			description:    "the float 42.120 as '42.12",
			value:          42.120,
			expectedString: "42.12",
		},
		{
			description:    "the error new error as 'new error'",
			value:          errors.New("new error"),
			expectedString: "new error",
		},
		{
			description: "a struct with all fields and values shown'",
			value: struct {
				foo           int
				bar           string
				ExportedField string
			}{foo: 42, bar: "foobar", ExportedField: "exported"},
			expectedString: "{foo:42 bar:foobar ExportedField:exported}",
		},
	} {
		t.Run(fmt.Sprintf("parseValue should parse %s", tc.description), func(t *testing.T) {
			assert.Equal(t, tc.expectedString, parseValue(tc.value))
		})
	}
	t.Run("parseValue should correctly parse pointers", func(t *testing.T) {
		var foo int = 42
		pointerRepresentation := parseValue(&foo)
		assert.Equal(t, "0x", pointerRepresentation[:2])
	})
}
