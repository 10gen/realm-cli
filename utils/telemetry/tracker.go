package telemetry

import (
	"gopkg.in/segmentio/analytics-go.v3"
)

var (
	segmentWriteKey string //value will be injected at build-time
)

// REALMC-7243 same as the new cli Segment tracking
type Tracker interface {
	Track(event event)
	Close() error
}

type noopTracker struct{}

func (n *noopTracker) Track(event event) {}
func (n *noopTracker) Close() error      { return nil }

type segmentTracker struct {
	client analytics.Client
}

func newSegmentTracker() Tracker {
	if len(segmentWriteKey) == 0 {
		return &noopTracker{}
	}
	client := analytics.New(segmentWriteKey)
	return &segmentTracker{client}
}

func (s *segmentTracker) Track(event event) {
	properties := make(map[string]interface{}, len(event.data)+2)
	properties[eventDataKeyCommand] = event.command
	properties[eventDataKeyExecutionID] = event.executionID

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
		return
	}
}

func (s *segmentTracker) Close() error {
	return s.client.Close()
}
