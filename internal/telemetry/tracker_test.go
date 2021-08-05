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

func TestStdoutTracker(t *testing.T) {
	t.Run("should create an stdout tracker and should print the tracking information to stdout", func(t *testing.T) {
		tracker := stdoutTracker{}
		testEvent := createEvent(EventTypeCommandComplete, nil, testCommand)

		r, w, resetStdout := mockStdoutSetup(t)
		defer resetStdout()

		tracker.Track(testEvent)

		assert.Nil(t, w.Close())

		out, err := ioutil.ReadAll(r)
		assert.Nil(t, err)
		assert.Equal(t, "command: COMMAND_COMPLETE[]\n", string(out))
	})
}

func TestSegmentTracker(t *testing.T) {
	t.Run("should create the segment tracker and should print the tracking information to the logger", func(t *testing.T) {
		client := &mockSegmentClient{}

		tracker := segmentTracker{}
		tracker.client = client

		tracker.Track(createEvent(EventTypeCommandError, EventDataError(errors.New("error")), testCommand))

		assert.Equal(t, 1, len(client.calls))

		track, ok := client.calls[0].(analytics.Track)
		assert.True(t, ok, "expected client call to be a track")

		assert.Equal(t, testID, track.MessageId)
		assert.Equal(t, testUserID, track.UserId)
		assert.Equal(t, testTime, track.Timestamp)
		assert.Equal(t, string(EventTypeCommandError), track.Event)
		assert.Equal(t, errors.New("error"), track.Properties[eventDataKeyError])
		assert.Equal(t, "error", track.Properties[eventDataKeyErrorMessage])
		assert.Equal(t, testCommand, track.Properties[eventDataKeyCommand])
		assert.Equal(t, testExecutionID, track.Properties[eventDataKeyExecutionID])
		assert.Equal(t, testVersion, track.Properties[eventDataKeyVersion])
	})

	t.Run("should capture the error in the logger passed in", func(t *testing.T) {
		client := &mockFailureSegmentClient{}
		tracker := &segmentTracker{}
		tracker.client = client

		r, w, resetStdout := mockStdoutSetup(t)
		defer resetStdout()

		tracker.Track(createEvent(EventTypeCommandError, EventDataError(errors.New("error")), testCommand))

		assert.Nil(t, w.Close())

		out, err := ioutil.ReadAll(r)
		assert.Nil(t, err)
		assert.Equal(t, "", string(out))
	})
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
