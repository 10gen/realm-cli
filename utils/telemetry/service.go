package telemetry

import (
	"time"

	"github.com/10gen/realm-cli/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service tracks telemetry events
type Service struct {
	NoTelemetry bool
	Tracker     Tracker
	userID      string
	executionID string
	command     string
}

// Setup sets up the tracker for this service
func (s *Service) Setup(command string) {
	s.command = command
	s.executionID = primitive.NewObjectID().Hex()
	if !s.NoTelemetry {
		s.Tracker = newSegmentTracker()
	}
}

// SetUser sets the userID for this service to track
func (s *Service) SetUser(userID string) {
	s.userID = userID
}

// TrackEvent tracks the event based on the tracker
func (s *Service) TrackEvent(eventType EventType, data ...EventData) {
	s.Tracker.Track(event{
		id:          primitive.NewObjectID().Hex(),
		eventType:   eventType,
		userID:      s.userID,
		time:        time.Now(),
		executionID: s.executionID,
		version:     utils.CLIVersion,
		command:     s.command,
		data:        data,
	})
}

// Close shuts down the Service
func (s *Service) Close() {
	s.Tracker.Close()
}
