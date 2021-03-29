package telemetry

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/mitchellh/cli"
)

// REALMC-7243 same as the new cli Segment tracking

type Service struct {
	userID      string
	executionID string
	command     string
	tracker     Tracker
}

func NewService(mode Mode, userID string, command string, ui cli.Ui) *Service {
	if !isValid(mode){
		return nil
	}
	service := Service{
		userID:      userID,
		executionID: primitive.NewObjectID().Hex(),
		command:     command,
	}

	switch mode {
	case ModeOn, ModeEmpty:
		service.tracker = newSegmentTracker(ui)
	case ModeOff:
		service.tracker = &noopTracker{}
	default:
		service.tracker = &noopTracker{}
	}
	return &service
}

func (s *Service) TrackEvent(eventType EventType, data ...EventData) {
	s.tracker.Track(
		event{
			id:          primitive.NewObjectID().Hex(),
			eventType:   eventType,
			userID:      s.userID,
			time:        time.Now(),
			executionID: s.executionID,
			command:     s.command,
			data:        data,
		},
	)
}

func (s *Service) Close() error {
	return s.tracker.Close()
}
