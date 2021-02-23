package telemetry

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"time"
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

type segmentTracker struct{
	client analytics.Client
	logger *log.Logger
}

func newSegmentTracker(writeKey string, logger *log.Logger) (*segmentTracker, error){
	client, err := analytics.NewWithConfig(writeKey, analytics.Config{Endpoint: analytics.DefaultEndpoint})
	if err != nil {
		return nil, err
	}
	return &segmentTracker{client, logger}, err
}

// TODO(REALMC-7243): use Segment sdk to send events through client
func (tracker *segmentTracker) Track(event event) {
	if err := tracker.client.Enqueue(analytics.Track{
		MessageId:  primitive.NewObjectID().Hex(),
		Timestamp:  time.Now(),
		Event:      string(event.eventType),
		UserId:     event.userID,
		Properties: event.createPropertyMap(),
	}); err != nil {
		// TODO: REALMC-8240 Is Fatalf appropriate for this?
		tracker.logger.Fatalf("failed to send Segment event %q: %s", event.eventType, err)
	}
}
