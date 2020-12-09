package telemetry

import (
	"time"
)

// Event is a telemetry event
type event struct {
	id          string
	eventType   EventType
	userID      string
	time        time.Time
	executionID string
	command     string
	data        []EventData
}

// EventData holds additional Event information
type EventData struct {
	Key   EventDataKey
	Value interface{}
}

// EventType is a cli event type
type EventType string

// set of supported cli event types
const (
	EventTypeCommandStart    EventType = "COMMAND_START"
	EventTypeCommandComplete EventType = "COMMAND_COMPLETE"
	EventTypeCommandError    EventType = "COMMAND_ERROR"
)

// EventDataKey used to pass data into the Event.data map
type EventDataKey string

// set of Data Keys
const (
	EventDataKeyErr EventDataKey = "err"
)
