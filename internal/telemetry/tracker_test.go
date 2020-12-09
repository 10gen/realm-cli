package telemetry

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"github.com/google/go-cmp/cmp"
)

func TestServiceConstructor(t *testing.T) {
	assert.RegisterOpts(reflect.TypeOf(&service{}), cmp.AllowUnexported(service{}))

	t.Run(fmt.Sprintf("Should create the expected Service"), func(t *testing.T) {
		constructedService := NewService(ModeOn, "userID", "executionID", "command")
		expectedService := &service{
			userID:      "userID",
			executionID: "executionID",
			command:     "command",
			tracker:     &segmentTracker{},
		}
		assert.Equal(t, expectedService, constructedService)
	})
}

func TestEventTracking(t *testing.T) {

	t.Run(fmt.Sprintf("Should track the expected event"), func(t *testing.T) {
		trackerService := NewService(modeTest, "userID", "executionID", "command")
		trackerService.TrackEvent(EventTypeCommandError, EventData{Key: EventDataKeyErr, Value: fmt.Errorf("error")})
		trackerService.TrackEvent(EventTypeCommandError, EventData{Key: EventDataKeyErr, Value: fmt.Errorf("error")})
		trackedEvent := trackerService.(*service).tracker.(*testTracker).lastTrackedEvent
		expectedEvent := &event{
			eventType:   EventTypeCommandError,
			userID:      "userID",
			executionID: "executionID",
			command:     "command",
			data:        []EventData{{Key: EventDataKeyErr, Value: fmt.Errorf("error")}}}

		assert.Equal(t, expectedEvent.eventType, trackedEvent.eventType)
		assert.Equal(t, expectedEvent.userID, trackedEvent.userID)
		assert.Equal(t, expectedEvent.executionID, trackedEvent.executionID)
		assert.Equal(t, expectedEvent.command, trackedEvent.command)
		assert.Equal(t, len(expectedEvent.data), len(trackedEvent.data))
		assert.Equal(t, expectedEvent.data[0].Key, trackedEvent.data[0].Key)
		assert.Equal(t, expectedEvent.data[0].Value, trackedEvent.data[0].Value)
	})
}
