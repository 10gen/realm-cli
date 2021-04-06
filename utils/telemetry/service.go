package telemetry

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// REALMC-7243 same as the new cli Segment tracking

type Service struct {
	userID      string
	executionID string
	command     string
	tracker     Tracker
}

func (s *Service) SetFields(telemetryOn bool, userID string, command string) {
	s.userID = userID
	s.executionID = primitive.NewObjectID().Hex()
	s.command = command

	if telemetryOn {
		s.tracker = newSegmentTracker()
	} else {
		s.tracker = &noopTracker{}
	}
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
