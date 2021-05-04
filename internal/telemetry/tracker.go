package telemetry

import (
	"errors"
	"fmt"

	"gopkg.in/segmentio/analytics-go.v3"
)

var (
	segmentWriteKey = "" // value will be injected at build-time
)

// Tracker is a telemetry event tracker
type Tracker interface {
	Track(event event)
	Close()
}

type stdoutTracker struct{}

func (tracker stdoutTracker) Track(event event) {
	fmt.Printf("%s: %s%v\n", event.command, event.eventType, event.data)
}

func (tracker stdoutTracker) Close() {}

type segmentTracker struct {
	client analytics.Client
}

func newSegmentTracker() (Tracker, error) {
	if len(segmentWriteKey) == 0 {
		return nil, errors.New("no write key")
	}

	client, err := analytics.NewWithConfig(segmentWriteKey, analytics.Config{Logger: segmentNoopLogger{}})
	if err != nil {
		return nil, err
	}

	return &segmentTracker{client}, nil
}

func (tracker *segmentTracker) Track(event event) {
	properties := make(map[string]interface{}, len(event.data)+3)
	properties[eventDataKeyCommand] = event.command
	properties[eventDataKeyExecutionID] = event.executionID
	properties[eventDataKeyVersion] = event.version

	for _, datum := range event.data {
		properties[datum.Key] = datum.Value
	}
	if err := tracker.client.Enqueue(analytics.Track{
		MessageId:  event.id,
		Timestamp:  event.time,
		Event:      string(event.eventType),
		UserId:     event.userID,
		Properties: properties,
	}); err != nil {
		return // do nothing
	}
}

func (tracker *segmentTracker) Close() {
	// flush the client on close so that all queued events are sent
	tracker.client.Close()
}

type segmentNoopLogger struct {
}

// Logf is a no-op implementation of the Segment logger's log function
func (l segmentNoopLogger) Logf(format string, args ...interface{}) {}

// Errorf is a no-op implementation of the Segment logger's error function
func (l segmentNoopLogger) Errorf(format string, args ...interface{}) {}
