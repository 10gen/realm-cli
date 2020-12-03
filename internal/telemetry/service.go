package telemetry

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service tracks telemetry events
type Service struct {
	userID      string
	executionID string
	command     string
	tracker     Tracker
}

// NewService creates a new telemetry service
func NewService(mode Mode, userID string, command string) *Service {
	service := Service{
		userID:      userID,
		command:     command,
		executionID: primitive.NewObjectID().Hex(),
	}

	switch mode {
	case ModeOff:
		service.tracker = &noopTracker{}
	case ModeNil, ModeOn:
		service.tracker = &segmentTracker{}
	case ModeStdout:
		service.tracker = &stdoutTracker{}
	}

	return &service
}

// TrackEvent tracks events
func (service *Service) TrackEvent(eventType EventType, data ...EventData) {
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
