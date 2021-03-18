package telemetry

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/segmentio/analytics-go.v3"
)

var (
	segmentWriteKey = "" // value will be injected at build-time
)

// Tracker is a telemetry event tracker
type Tracker interface {
	Track(event event)
}

type noopTracker struct{}

func (tracker *noopTracker) Track(event event) {}

type stdoutTracker struct{}

func (tracker *stdoutTracker) Track(event event) {
	fmt.Printf(
		"%s UTC TELEM %s: %s%v\n",
		event.time.In(time.UTC).Format("15:04:05"),
		event.command,
		event.eventType,
		event.data,
	)
}

type segmentTracker struct {
	client analytics.Client
	logger *log.Logger
}

func newSegmentTracker(logger *log.Logger) Tracker {
	if len(segmentWriteKey) == 0 {
		return &noopTracker{}
	}
	client := analytics.New(segmentWriteKey)
	return &segmentTracker{client, logger}
}

func (tracker *segmentTracker) Track(event event) {
	properties := make(map[string]interface{}, len(event.data)+2)
	properties[eventDataKeyCommand] = event.command
	properties[eventDataKeyExecutionID] = event.executionID

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
		tracker.logger.Printf("failed to send Segment event %q: %s", event.eventType, err)
	}
}
