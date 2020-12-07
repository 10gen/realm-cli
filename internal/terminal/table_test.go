package terminal

import (
	"fmt"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/fatih/color"
)

func TestNewTable(t *testing.T) {
	for _, tc := range []struct {
		description    string
		header         []string
		data           []map[string]interface{}
		expectedHeader []string
		expectedData   []map[string]string
		expectedWidths map[string]int
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
			description:    "Should return a table with only headers if there is one and no data",
			header:         []string{"should", "show", "up", "only"},
			expectedHeader: []string{"should", "show", "up", "only"},
			expectedData:   make([]map[string]string, 0),
			expectedWidths: map[string]int{
				"should": 6,
				"show":   4,
				"up":     2,
				"only":   4,
			},
		},
		{
			description: "Should return a table if there is both a header and data",
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
			expectedHeader: []string{"test"},
			expectedData: []map[string]string{
				{
					"test": "should show up",
				},
			},
			expectedWidths: map[string]int{
				"test": 14,
			},
		},
	} {
		t.Run(tc.description, func(*testing.T) {
			table := newTable(tc.header, tc.data)
			assert.Equal(t, tc.expectedWidths, table.columnWidths)
			assert.Equal(t, tc.expectedData, table.data)
			assert.Equal(t, tc.expectedHeader, table.headers)
		})
	}
}

func TestTableMessage(t *testing.T) {
	t.Run("Should return an empty string and an error if the table has no headers", func(t *testing.T) {
		table := newTable(nil, nil)
		message, err := table.Message()
		assert.Equal(t, message, "")
		assert.Equal(t, err.Error(), "cannot create a table without headers")
	})

	for _, tc := range []struct {
		description     string
		header          []string
		data            []map[string]interface{}
		expectedMessage string
	}{
		{
			description: "Should print only a header for a table with no data",
			header:      []string{"header", "only", "no", "data"},
			expectedMessage: "\n" + fmt.Sprintf("%s  %s  %s  %s",
				[]interface{}{
					color.New(color.Bold).SprintFunc()("header"),
					color.New(color.Bold).SprintFunc()("only"),
					color.New(color.Bold).SprintFunc()("no"),
					color.New(color.Bold).SprintFunc()("data"),
				}...,
			) + `
------  ----  --  ----
`,
		},

		{
			description: "Should return correctly formatted values in the table, not create new columns, and not print empty rows",
			header:      []string{"arrays", "floats", "ints", "maps/objects", "strings", "sparse"},
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
			expectedMessage: "\n" + fmt.Sprintf("%s                  %s  %s  %s               %s        %s",
				[]interface{}{
					color.New(color.Bold).SprintFunc()("arrays"),
					color.New(color.Bold).SprintFunc()("floats"),
					color.New(color.Bold).SprintFunc()("ints"),
					color.New(color.Bold).SprintFunc()("maps/objects"),
					color.New(color.Bold).SprintFunc()("strings"),
					color.New(color.Bold).SprintFunc()("sparse"),
				}...,
			) + `
----------------------  ------  ----  -------------------------  -------------  ------
[1 test this]           12.34   1     {1234 12.345}              tester string        
[1 2 3 4]               12.34   1     map[test:1 this:2]         test           hello 
[1 2.3 4.5555555555 6]  123.34  -2    map[what happens:[1 2 3]]  hello                `,
		},
	} {
		t.Run(tc.description, func(*testing.T) {
			table := newTable(tc.header, tc.data)
			message, err := table.Message()
			assert.Nil(t, err)
			assert.Equal(t, message, tc.expectedMessage)
		})
	}
}

func TestTableMessageNoBold(t *testing.T) {
	for _, tc := range []struct {
		description     string
		header          []string
		data            []map[string]interface{}
		expectedMessage string
	}{
		{
			description: "Should not bold the headers if color is off",
			header:      []string{"should", "not", "be", "bold"},
			expectedMessage: `
should  not  be  bold
------  ---  --  ----
`,
		},
		{
			description: "Should not bold the headers if color is off even if there is data",
			header:      []string{"should", "not", "be", "bold"},
			data: []map[string]interface{}{
				{
					"should": 123,
					"be":     "not bold!",
				},
			},
			expectedMessage: `
should  not  be         bold
------  ---  ---------  ----
123          not bold!      `,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			color.NoColor = true
			defer func() { color.NoColor = false }()

			table := newTable(tc.header, tc.data)
			message, err := table.Message()
			assert.Nil(t, err)
			assert.Equal(t, message, tc.expectedMessage)
		})
	}
}

func TestTablePayload(t *testing.T) {
	t.Run("Payload should return an error with a table without a header", func(t *testing.T) {
		table := newTable(nil, nil)
		payloadKeys, payloadData, err := table.Payload()
		assert.Nil(t, payloadKeys)
		assert.Nil(t, payloadData)
		assert.Equal(t, err.Error(), "cannot create a table without headers")
	})

	t.Run("Payload should work with a valid table", func(t *testing.T) {
		header := []string{"test", "this", "data"}
		data := []map[string]interface{}{
			{"test": 123, "this": "456", "data": []string{"7", "8 9", "10!"}},
		}
		table := newTable(header, data)
		payloadKeys, payloadData, err := table.Payload()
		assert.Nil(t, err)
		assert.Equal(t, payloadKeys, []string{logFieldData, logFieldHeaders})
		assert.Equal(t, payloadData[logFieldHeaders], header)

		expectedData := []map[string]string{
			{
				"test": "123",
				"this": "456",
				"data": "[7 8 9 10!]",
			},
		}
		assert.Equal(t, expectedData, payloadData[logFieldData])
	})
}
