package telemetry

import (
	"log"
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
func NewService(mode Mode, userID string, logger *log.Logger, command string) *Service {
	service := Service{
		userID:      userID,
		command:     command,
		executionID: primitive.NewObjectID().Hex(),
	}

	switch mode {
	case ModeOff:
		service.tracker = &noopTracker{}
	case ModeEmpty, ModeOn:
		service.tracker = newSegmentTracker(logger)
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
		data:        data,
	})
}

// Close shuts down the Service
func (service Service) Close() error {
	return service.tracker.Close()
}
