package terminal

import (
	"fmt"
	"testing"
)

func TestNewList(t *testing.T) {
	for _, tc := range []struct {
		description  string
		message      string
		data         []interface{}
		expectedList list
	}{
		{
			description: "Should return a list",
			message:     "a list message",
			data: []interface{}{
				"should show up",
				1,
				1.000,
				1.234567890123,
				[]string{"1", "2", "3"},
				nil,
			},
			expectedList: list{
				"a list message",
				[]string{
					"should show up",
					"1",
					"1",
					"1.234567890123",
					"[1 2 3]",
				},
				14,
			},
		},
	} {
		t.Run(tc.description, func(*testing.T) {
			list := newList(tc.message, tc.data)
			fmt.Printf(list.Message())
			//assert.Equal(t, tc.expectedList, list)
		})
	}
}
