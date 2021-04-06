package telemetry

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service tracks telemetry events
type Service struct {
	userID      string
	command     string
	version     string
	executionID string
	tracker     Tracker
}

// NewService creates a new telemetry service
func NewService(mode Mode, userID, command, version string) *Service {
	service := Service{
		userID:      userID,
		command:     command,
		version:     version,
		executionID: primitive.NewObjectID().Hex(),
	}

	switch mode {
	case ModeOff:
		service.tracker = &noopTracker{}
	case ModeEmpty, ModeOn:
		service.tracker = newSegmentTracker()
	case ModeStdout:
		service.tracker = &stdoutTracker{}
	}

	return &service
}

// TrackEvent tracks events
func (service *Service) TrackEvent(eventType EventType, data ...EventData) {
	service.tracker.Track(event{
		id:          primitive.NewObjectID().Hex(),
		eventType:   eventType,
		userID:      service.userID,
		time:        time.Now(),
		executionID: service.executionID,
		command:     service.command,
		version:     service.version,
		data:        data,
	})
}

// Close shuts down the Service
func (service Service) Close() {
	service.tracker.Close()
}
