package terminal

import (
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/google/go-cmp/cmp"
)

func TestList(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(list{}), cmp.AllowUnexported(list{}))
	message := "a list message"
	data := []interface{}{
		"should show up",
		1,
		1.000,
		1.234567890123,
		[]string{"1", "2", "3"},
		nil,
	}

	t.Run("newList() should work", func(*testing.T) {
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
		}
		assert.Equal(t, expectedList, newList(message, data))
	})

	t.Run("list.Message() should work", func(*testing.T) {
		list := newList(message, data)
		expectedMessage := `a list message
  should show up
  1
  1
  1.234567890123
  [1 2 3]
  
`
		message, err := list.Message()
		assert.Nil(t, err)
		assert.Equal(t, expectedMessage, message)
	})

	t.Run("list.Payload() should work", func(*testing.T) {
		list := newList(message, data)
		expectedPayloadKeys := []string{"message", "data"}
		expectedPayloadData := map[string]interface{}{
			"message": message,
			"data": []string{
				"should show up",
				"1",
				"1",
				"1.234567890123",
				"[1 2 3]",
				"",
			},
		}
		payloadKeys, payloadData, err := list.Payload()
		assert.Nil(t, err)
		assert.Equal(t, expectedPayloadKeys, payloadKeys)
		assert.Equal(t, expectedPayloadData, payloadData)
	})
}
