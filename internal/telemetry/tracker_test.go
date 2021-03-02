package telemetry

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"gopkg.in/segmentio/analytics-go.v3"
)

type testClient struct {
	calls []interface{}
}

func (client *testClient) Enqueue(track analytics.Message) error {
	client.calls = append(client.calls, track)
	return nil
}

func (client *testClient) Close() error {
	return nil
}

const (
	testID          = "id123"
	testUserID      = "user123"
	testExecutionID = "execution123"
)

var testTime = time.Date(2021, 1, 2, 3, 4, 5, 6, time.UTC)

func TestNoopTracker(t *testing.T) {
	t.Run("should create a noop tracker and should not do anything when track is called", func(t *testing.T) {
		tracker := noopTracker{}
		testEvent := createEvent(EventTypeCommandStart, nil, "someCommand")
		testTrackStdoutOutput(t, &tracker, testEvent, "")
	})
}

func TestStdoutTracker(t *testing.T) {
	t.Run("should create an stdout tracker and should print the tracking information to stdout", func(t *testing.T) {
		tracker := stdoutTracker{}
		testTrackStdoutOutput(
			t,
			&tracker,
			createEvent(EventTypeCommandComplete, nil, "someCommand"),
			"03:04:05 UTC TELEM someCommand: COMMAND_COMPLETE[]\n",
		)
	})
}

func TestSegmentTracker(t *testing.T) {
	client := &testClient{}
	t.Run("should create the segment tracker and should print the tracking information to the logger", func(t *testing.T) {
		tracker := segmentTracker{}
		tracker.client = client
		tracker.Track(createEvent(EventTypeCommandError, []EventData{{Key: EventDataKeyErr, Value: "Something"}}, "someCommand"))

		testClientResults := tracker.client.(*testClient)
		expectedTrack := analytics.Track{
			MessageId: testID,
			UserId:    testUserID,
			Timestamp: testTime,
			Event:     string(EventTypeCommandError),
			Properties: map[string]interface{}{
				string(EventDataKeyErr): "Something",
				string(eventDataKeyCmd): "someCommand",
				string(eventDataKeyExecutionID): testExecutionID,
			},
		}

		actualTrack := testClientResults.calls[0].(analytics.Track)
		assert.NotNil(t, actualTrack)
		assert.Equal(t, []interface{}{expectedTrack}, testClientResults.calls)
	})
}

func testTrackStdoutOutput(t *testing.T, tracker Tracker, event event, expected string) {
	t.Helper()

	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()
	r, w, err := os.Pipe()
	assert.Nil(t, err)
	os.Stdout = w

	tracker.Track(event)

	assert.Nil(t, w.Close())

	out, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(out))
}

func createEvent(eventType EventType, data []EventData, command string) event {
	return event{
		id:          testID,
		eventType:   eventType,
		userID:      testUserID,
		time:        testTime,
		executionID: testExecutionID,
		command:     command,
		data:        data,
	}
}
