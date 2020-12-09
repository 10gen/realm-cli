package telemetry

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"

	"github.com/google/go-cmp/cmp"
)

func TestEventConstructor(t *testing.T) {
	assert.RegisterOpts(
		reflect.TypeOf(map[DataKey]interface{}{}),
		cmp.Comparer(func(x, y error) bool {
			if x == nil || y == nil {
				return x == nil && y == nil
			}
			return x.Error() == y.Error()
		}))

	for _, tc := range []struct {
		ctor          string
		event         Event
		expectedEvent Event
	}{
		{
			ctor:  "NewCommandStartEvent",
			event: NewCommandStartEvent("command"),
			expectedEvent: Event{
				eventType: EventTypeCommandStart,
				command:   "command",
				data:      map[DataKey]interface{}{},
			},
		},
		{
			ctor:  "NewCommandCompleteEvent",
			event: NewCommandCompleteEvent("command"),
			expectedEvent: Event{
				eventType: EventTypeCommandStart,
				command:   "command",
				data:      map[DataKey]interface{}{},
			},
		},
		{
			ctor:  "NewCommandErrorEvent",
			event: NewCommandErrorEvent("command", fmt.Errorf("error")),
			expectedEvent: Event{
				eventType: EventTypeCommandStart,
				userID:    "user",
				command:   "command",
				data: map[DataKey]interface{}{
					DataKeyErr: fmt.Errorf("error"),
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s should create the expected Event", tc.ctor), func(t *testing.T) {
			assert.Equal(t, tc.expectedEvent.eventType, tc.event.eventType)
			assert.Equal(t, tc.expectedEvent.command, tc.event.command)
			assert.Equal(t, tc.expectedEvent.data, tc.event.data)
		})
	}
}
