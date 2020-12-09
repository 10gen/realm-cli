package telemetry

import (
	"fmt"
)

// Tracker logs events
type Tracker interface {
	Track(event Event)
}

type tracker struct {
	userID          string
	executionID     string
	telemetryClient telemetryClient
}

// NewTracker creates a new tracker of type mode with userID and executionID appended to each event
func NewTracker(mode Mode, userID string, executionID string) Tracker {
	tracker := tracker{userID: userID, executionID: executionID}
	switch mode {
	case ModeOn, ModeNil:
		tracker.telemetryClient = &segmentClient{}
	case ModeStdout:
		tracker.telemetryClient = &stdoutClient{}
	default:
		tracker.telemetryClient = &noopClient{}
	}
	return &tracker
}

func (tracker *tracker) Track(event Event) {
	event.userID = tracker.userID
	event.executionID = tracker.executionID
	tracker.telemetryClient.track(event)
}

type telemetryClient interface {
	track(event Event)
}

type noopClient struct{}

func (client *noopClient) track(event Event) {}

type stdoutClient struct{}

func (client *stdoutClient) track(event Event) {
	fmt.Printf("tracking: %v\n", event)
}

type segmentClient struct{}

func (client *segmentClient) track(event Event) {
	fmt.Printf("tracking: %v\n", event)
}
