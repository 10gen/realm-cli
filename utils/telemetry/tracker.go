package telemetry

import (
	"gopkg.in/segmentio/analytics-go.v3"
)

var (
	segmentWriteKey string //value will be injected at build-time
)

// Tracker is a telemetry event tracker
type Tracker interface {
	Track(event event)
	Close()
}

type NoopTracker struct{}

func (n *NoopTracker) Track(event event) {}
func (n *NoopTracker) Close()            {}

type segmentTracker struct {
	client analytics.Client
}

func newSegmentTracker() Tracker {
	if len(segmentWriteKey) == 0 {
		return &NoopTracker{}
	}
	client := analytics.New(segmentWriteKey)
	return &segmentTracker{client}
}

func (s *segmentTracker) Track(event event) {
	properties := make(map[string]interface{}, len(event.data)+3)
	properties[eventDataKeyCommand] = event.command
	properties[eventDataKeyExecutionID] = event.executionID
	properties[eventDataKeyVersion] = event.version

	for _, datum := range event.data {
		properties[datum.Key] = datum.Value
	}

	err := s.client.Enqueue(analytics.Track{
		MessageId:  event.id,
		Timestamp:  event.time,
		Event:      string(event.eventType),
		UserId:     event.userID,
		Properties: properties,
	})
	if err != nil {
		return // do nothing
	}
}

func (s *segmentTracker) Close() {
	// flush the client on close so that all queued events are sent
	s.client.Close()
}
