package telemetry

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

func TestNewService(t *testing.T) {
	t.Run("Should create the expected Service", func(t *testing.T) {
		service := NewService(ModeStdout, "userID", nil, "command")

		assert.Equal(t, "command", service.command)
		assert.True(t, service.executionID != "", "service execution id must not be blank")
		assert.Equal(t, "userID", service.userID)
		assert.NotNil(t, service.tracker)
	})

	createSegmentFn := func() *Service {
		return NewService(ModeOn, "userID", log.New(os.Stdout, "LogPrefix ", log.Lmsgprefix), "command")
	}
	t.Run("Should create a segment tracking service if the segmentWriteKey is there", func(t *testing.T) {
		swk := segmentWriteKey
		defer func() { segmentWriteKey = swk }()

		segmentWriteKey = "testing"
		testServiceStdoutput(t, createSegmentFn, "")

		service := createSegmentFn()

		assert.Equal(t, "command", service.command)
		assert.True(t, service.executionID != "", "service execution id must not be blank")
		assert.Equal(t, "userID", service.userID)
		assert.NotNil(t, service.tracker)
	})

	t.Run("Should disable the service if the segmentWriteKey is empty", func(t *testing.T) {
		swk := segmentWriteKey
		defer func() { segmentWriteKey = swk }()

		segmentWriteKey = ""
		testServiceStdoutput(t, createSegmentFn, "LogPrefix unable to connect to Segment due to missing key, CLI telemetry will be disabled\n")
		service := createSegmentFn()

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
func testServiceStdoutput(t *testing.T, createFn func() *Service, expected string) {
	t.Helper()

	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()
	r, w, err := os.Pipe()
	assert.Nil(t, err)
	os.Stdout = w

	createFn()

	assert.Nil(t, w.Close())

	out, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(out))
}
