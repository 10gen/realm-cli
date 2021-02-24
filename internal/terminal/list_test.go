package terminal

import (
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/google/go-cmp/cmp"
)

func TestList(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(list{}), cmp.AllowUnexported(list{}))

	t.Run("newList should create a list with its data parsed into strings", func(*testing.T) {
		testMessage := "a list message"
		testData := []interface{}{
			"should show up",
			1,
			1.000,
			1.234567890123,
			[]string{"1", "2", "3"},
			nil,
		}
		testList := newList(testMessage, testData, false)
		expectedList := list{
			"a list message",
			[]string{
				"should show up",
				"1",
				"1",
				"1.234567890123",
				"[1 2 3]",
				"",
			},
			false,
		}
		assert.Equal(t, expectedList, testList)

		t.Run("And Message should display a properly formatted list", func(*testing.T) {
			message, err := testList.Message()
			assert.Nil(t, err)
			expectedMessage := `a list message
  should show up
  1
  1
  1.234567890123
  [1 2 3]
  `
			assert.Equal(t, expectedMessage, message)
		})

		t.Run("And Payload should create a payload representation of the list", func(*testing.T) {
			payloadKeys, payloadData, err := testList.Payload()
			assert.Nil(t, err)
			expectedPayloadKeys := []string{"message", "data"}
			expectedPayloadData := map[string]interface{}{
				"message": testMessage,
				"data": []string{
					"should show up",
					"1",
					"1",
					"1.234567890123",
					"[1 2 3]",
					"",
				},
			}
			assert.Equal(t, expectedPayloadKeys, payloadKeys)
			assert.Equal(t, expectedPayloadData, payloadData)
		})
	})

	t.Run("Message should display single item lists the same line as the message with consolidated set to true", func(*testing.T) {
		testList := newList(MsgReferenceLinks, []interface{}{"https://mongodb.com"}, true)

		message, err := testList.Message()
		assert.Nil(t, err)
		assert.Equal(t, "For more information: https://mongodb.com", message)
	})

	t.Run("Message should display single item lists the same line as the message with consolidated set to false", func(*testing.T) {
		testList := newList(MsgReferenceLinks, []interface{}{"https://mongodb.com"}, false)

		message, err := testList.Message()
		assert.Nil(t, err)
		assert.Equal(t, "For more information\n  https://mongodb.com", message)
	})
}
