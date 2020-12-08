package telemetry

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/google/go-cmp/cmp"
)

func TestEventConstructor(t *testing.T) {
	//I agree, we should clean this up :/
	assert.RegisterOpts(
		reflect.TypeOf(map[DataKey]interface{}{}),
		cmp.Comparer(func(x, y error) bool {
			if x == nil || y == nil {
				return x == nil && y == nil
			}
			return x.Error() == y.Error()
		}))

	ConfigureEvents("user")

	for _, tc := range []struct {
		ctor          string
		event         *Event
		expectedEvent *Event
	}{
		{
			ctor:  "NewCommandStartEvent",
			event: NewCommandStartEvent("command"),
			expectedEvent: &Event{
				Type:   EventTypeCommandStart,
				UserID: "user",
				Data: map[DataKey]interface{}{
					DataKeyCommand:     "command",
					DataKeyExecutionID: executionID,
				},
			},
		},
		{
			ctor:  "NewCommandCompleteEvent",
			event: NewCommandCompleteEvent("command"),
			expectedEvent: &Event{
				Type:   EventTypeCommandStart,
				UserID: "user",
				Data: map[DataKey]interface{}{
					DataKeyCommand:     "command",
					DataKeyExecutionID: executionID,
				},
			},
		},
		{
			ctor:  "NewCommandErrorEvent",
			event: NewCommandErrorEvent("command", fmt.Errorf("error")),
			expectedEvent: &Event{
				Type:   EventTypeCommandStart,
				UserID: "user",
				Data: map[DataKey]interface{}{
					DataKeyCommand:     "command",
					DataKeyExecutionID: executionID,
					DataKeyErr:         fmt.Errorf("error"),
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s should create the expected Event", tc.ctor), func(t *testing.T) {
			assert.Equal(t, tc.expectedEvent.Type, tc.event.Type)
			assert.Equal(t, tc.expectedEvent.UserID, tc.event.UserID)
			assert.Equal(t, tc.expectedEvent.Data, tc.event.Data)
		})
	}
}
