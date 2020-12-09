package telemetry

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service tracks events
type Service interface {
	TrackEvent(eventType EventType, data ...EventData)
}

type service struct {
	userID      string
	executionID string
	command     string
	tracker     tracker
}

// NewService creates a new service of type mode with userID and
// executionID appended to each event it tracks
func NewService(mode Mode, userID string, executionID string, command string) Service {
	service := service{userID: userID, executionID: executionID, command: command}
	switch mode {
	case ModeOn, ModeNil:
		service.tracker = &segmentTracker{}
	case ModeStdout:
		service.tracker = &stdoutTracker{}
	case modeTest:
		service.tracker = &testTracker{}
	default:
		service.tracker = &noopTracker{}
	}
	return &service
}

// TrackEvents tracks events
func (service *service) TrackEvent(eventType EventType, data ...EventData) {
	service.tracker.track(event{
		id:          primitive.NewObjectID().Hex(),
		eventType:   eventType,
		userID:      service.userID,
		time:        time.Now(),
		executionID: service.executionID,
		command:     service.command,
		data:        data,
	})
}

type tracker interface {
	track(event event)
}

type noopTracker struct{}

func (tracker *noopTracker) track(event event) {}

type stdoutTracker struct{}

func (tracker *stdoutTracker) track(event event) {
	fmt.Printf("tracking: %v\n", event)
}

type segmentTracker struct{}

func (tracker *segmentTracker) track(event event) {
	fmt.Printf("tracking: %v\n", event)
}

type testTracker struct {
	lastTrackedEvent event
}

func (tracker *testTracker) track(event event) {
	tracker.lastTrackedEvent = event
}
