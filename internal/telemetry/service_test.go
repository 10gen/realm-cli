package telemetry

import (
	"errors"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestNewService(t *testing.T) {
	t.Run("Should create the expected Service", func(t *testing.T) {
		service := NewService(ModeOn, "userID",  nil, "command")
		assert.Equal(t, "command", service.command)
		assert.True(t, service.executionID != "", "service execution id must not be blank")
		assert.Equal(t, "userID", service.userID)
		assert.NotNil(t, service.tracker)
	})
}

func TestServiceTrackEvent(t *testing.T) {
	t.Run("Should track the expected event", func(t *testing.T) {
		tracker := &testTracker{}
		service := &Service{
			command:     "command",
			executionID: "executionID",
			userID:      "userID",
			tracker:     tracker,
		}

		service.TrackEvent(EventTypeCommandError, EventData{Key: EventDataKeyErr, Value: errors.New("error")})

		assert.Equal(t, EventTypeCommandError, tracker.lastTrackedEvent.eventType)
		assert.Equal(t, "command", tracker.lastTrackedEvent.command)
		assert.Equal(t, "executionID", tracker.lastTrackedEvent.executionID)
		assert.Equal(t, "userID", tracker.lastTrackedEvent.userID)
		assert.Equal(t, 1, len(tracker.lastTrackedEvent.data))
		assert.Equal(t, EventDataKeyErr, tracker.lastTrackedEvent.data[0].Key)
		assert.Equal(t, errors.New("error"), tracker.lastTrackedEvent.data[0].Value)
	})
}

type testTracker struct {
	lastTrackedEvent event
}

func (tracker *testTracker) Track(event event) {
	tracker.lastTrackedEvent = event
}
