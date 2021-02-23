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
func NewService(mode Mode, userID string, writeKey string, logger *log.Logger, command string) (*Service, error) {
	service := Service{
		userID:      userID,
		command:     command,
		executionID: primitive.NewObjectID().Hex(),
	}

	var tracker Tracker
	var err error
	switch mode {

	case ModeOff:
		tracker = &noopTracker{}
	case ModeEmpty, ModeOn:
		tracker, err = newSegmentTracker(writeKey, logger)
		if err != nil {
			tracker = &noopTracker{}
		}
	case ModeStdout:
		tracker = &stdoutTracker{}
	}
	service.tracker = tracker

	return &service, err
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
