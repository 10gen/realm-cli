package telemetry

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
)

const (
	testUser    = "userID"
	testCommand = "command"
	testVersion = "version"
	testXID     = "executionID"
)

func TestNewService(t *testing.T) {
	t.Run("Should create the expected Service", func(t *testing.T) {
		service := NewService(ModeStdout, testUser, testCommand, testVersion)

		assert.Equal(t, testCommand, service.command)
		assert.True(t, service.executionID != "", "service execution id must not be blank")
		assert.Equal(t, testUser, service.userID)
		assert.Equal(t, testVersion, service.version)
		assert.NotNil(t, service.tracker)
	})

	t.Run("Should create a segment tracking service if the segmentWriteKey is there", func(t *testing.T) {
		swk := segmentWriteKey
		defer func() { segmentWriteKey = swk }()

		segmentWriteKey = "testing"
		testServiceOutput(t, ModeOn, "")
	})
}

func TestServiceTrackEvent(t *testing.T) {
	t.Run("Should track the expected event", func(t *testing.T) {
		tracker := &testTracker{}
		service := &Service{
			command:     testCommand,
			userID:      testUser,
			version:     testVersion,
			executionID: testXID,
			tracker:     tracker,
		}

		service.TrackEvent(EventTypeCommandError, EventData{Key: EventDataKeyError, Value: errors.New("error")})

		assert.Equal(t, EventTypeCommandError, tracker.lastTrackedEvent.eventType)
		assert.Equal(t, testCommand, tracker.lastTrackedEvent.command)
		assert.Equal(t, testXID, tracker.lastTrackedEvent.executionID)
		assert.Equal(t, testUser, tracker.lastTrackedEvent.userID)
		assert.Equal(t, testVersion, tracker.lastTrackedEvent.version)
		assert.Equal(t, 1, len(tracker.lastTrackedEvent.data))
		assert.Equal(t, EventDataKeyError, tracker.lastTrackedEvent.data[0].Key)
		assert.Equal(t, errors.New("error"), tracker.lastTrackedEvent.data[0].Value)
	})
}

type testTracker struct {
	lastTrackedEvent event
}

func (tracker *testTracker) Track(event event) {
	tracker.lastTrackedEvent = event
}

func (tracker *testTracker) Close() {}

func newService(mode Mode) *Service {
	return NewService(mode, testUser, testCommand, testVersion)
}

func mockStdoutSetup(t *testing.T) (*os.File, *os.File, func()) {
	t.Helper()

	stdout := os.Stdout

	r, w, err := os.Pipe()
	assert.Nil(t, err)
	os.Stdout = w

	return r, w, func() { os.Stdout = stdout }
}

func testServiceOutput(t *testing.T, mode Mode, expected string) {
	t.Helper()

	r, w, resetStdout := mockStdoutSetup(t)
	defer resetStdout()

	newService(mode)

	assert.Nil(t, w.Close())

	out, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(out))
}
