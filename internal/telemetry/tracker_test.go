package telemetry

import (
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/10gen/realm-cli/internal/utils/test/assert"
	"gopkg.in/segmentio/analytics-go.v3"
)

type mockSegmentClient struct {
	calls []interface{}
}

func (client *mockSegmentClient) Enqueue(track analytics.Message) error {
	client.calls = append(client.calls, track)
	return nil
}

func (client *mockSegmentClient) Close() error {
	return nil
}

type mockFailureSegmentClient struct{}

func (client *mockFailureSegmentClient) Enqueue(track analytics.Message) error {
	return errors.New("failed to enqueue")
}

func (client *mockFailureSegmentClient) Close() error {
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
		testEvent := createEvent(EventTypeCommandStart, nil, testCommand)
		testTrackerOutput(t, &tracker, testEvent, "")
	})
}

func TestStdoutTracker(t *testing.T) {
	t.Run("should create an stdout tracker and should print the tracking information to stdout", func(t *testing.T) {
		tracker := stdoutTracker{}
		testEvent := createEvent(EventTypeCommandComplete, nil, testCommand)
		testTrackerOutput(t, &tracker, testEvent, "03:04:05 UTC TELEM command: COMMAND_COMPLETE[]\n")
	})
}

func TestSegmentTracker(t *testing.T) {
	t.Run("should create the segment tracker and should print the tracking information to the logger", func(t *testing.T) {
		client := &mockSegmentClient{}

		tracker := segmentTracker{}
		tracker.client = client

		tracker.Track(createEvent(EventTypeCommandError, []EventData{{Key: EventDataKeyError, Value: "Something"}}, testCommand))

		assert.Equal(t, []interface{}{analytics.Track{
			MessageId: testID,
			UserId:    testUserID,
			Timestamp: testTime,
			Event:     string(EventTypeCommandError),
			Properties: map[string]interface{}{
				EventDataKeyError:       "Something",
				eventDataKeyCommand:     testCommand,
				eventDataKeyExecutionID: testExecutionID,
				eventDataKeyVersion:     testVersion,
			},
		}}, client.calls)
	})

	t.Run("should capture the error in the logger passed in", func(t *testing.T) {
		client := &mockFailureSegmentClient{}
		tracker := &segmentTracker{}
		tracker.client = client

		r, w, resetStdout := mockStdoutSetup(t)
		defer resetStdout()

		tracker.Track(createEvent(
			EventTypeCommandError,
			[]EventData{{Key: EventDataKeyError, Value: "Something"}}, testCommand,
		))

		assert.Nil(t, w.Close())

		out, err := ioutil.ReadAll(r)
		assert.Nil(t, err)
		assert.Equal(t, "", string(out))
	})
}

func testTrackerOutput(t *testing.T, tracker Tracker, event event, expectedOutput string) {
	t.Helper()

	r, w, resetStdout := mockStdoutSetup(t)
	defer resetStdout()

	tracker.Track(event)

	assert.Nil(t, w.Close())

	out, err := ioutil.ReadAll(r)
	assert.Nil(t, err)
	assert.Equal(t, expectedOutput, string(out))
}

func createEvent(eventType EventType, data []EventData, command string) event {
	return event{
		id:          testID,
		eventType:   eventType,
		userID:      testUserID,
		time:        testTime,
		executionID: testExecutionID,
		command:     command,
		version:     testVersion,
		data:        data,
	}
}
