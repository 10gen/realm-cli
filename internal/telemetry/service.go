package telemetry

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service tracks telemetry events
type Service interface {
	TrackEvent(eventType EventType, data ...EventData)
	Close()
}

// NewService creates a new telemetry service
func NewService(mode Mode, userID, command, version string) Service {
	if mode == ModeOff {
		return noopService{}
	}

	var tracker Tracker
	switch mode {
	case ModeEmpty, ModeOn:
		t, err := newSegmentTracker()
		if err != nil {
			return noopService{}
		}
		tracker = t
	case ModeStdout:
		tracker = stdoutTracker{}
	}

	return &trackingService{
		userID:      userID,
		command:     command,
		version:     version,
		executionID: primitive.NewObjectID().Hex(),
		tracker:     tracker,
	}
}

type trackingService struct {
	userID      string
	command     string
	version     string
	executionID string
	tracker     Tracker
}

// TrackEvent tracks events
func (s *trackingService) TrackEvent(eventType EventType, data ...EventData) {
	s.tracker.Track(event{
		id:          primitive.NewObjectID().Hex(),
		eventType:   eventType,
		userID:      s.userID,
		time:        time.Now(),
		executionID: s.executionID,
		command:     s.command,
		version:     s.version,
		data:        data,
	})
}

// Close shuts down the Service
func (s trackingService) Close() {
	s.tracker.Close()
}

type noopService struct{}

func (s noopService) TrackEvent(eventType EventType, data ...EventData) {}
func (s noopService) Close()                                            {}
