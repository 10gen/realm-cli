package telemetry

import (
	"fmt"
	"github.com/mitchellh/cli"
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
	ui     cli.Ui
}

func newSegmentTracker(ui cli.Ui) Tracker {
	if len(segmentWriteKey) == 0 {
		return &noopTracker{}
	}
	client := analytics.New(segmentWriteKey)
	return &segmentTracker{client, ui}
}

func (s *segmentTracker) Track(event event) {
	properties := make(map[string]interface{}, len(event.data)+2)
	properties[eventDataKeyCommand] = event.command
	properties[eventDataKeyExecutionID] = event.executionID

	for _, datum := range event.data {
		properties[datum.Key] = datum.Value
	}
	if err := s.client.Enqueue(analytics.Track{
		MessageId:  event.id,
		Timestamp:  event.time,
		Event:      string(event.eventType),
		UserId:     event.userID,
		Properties: properties,
	}); err != nil {
		s.ui.Info(fmt.Sprintf("failed to send Segment event %q: %s", event.eventType, err))
	}
}

func (s *segmentTracker) Close() error {
	return s.client.Close()
}
