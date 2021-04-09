package telemetry

import (
	"time"
)

type event struct {
	id          string
	eventType   EventType
	userID      string
	time        time.Time
	executionID string
	command     string
	version     string
	data        []EventData
}

// EventData outlines a generic structure for tracking events
type EventData struct {
	Key   string
	Value interface{}
}

// EventType signifies where in the command this event occurred
type EventType string

// set of supported EventTypes
const (
	EventTypeCommandStart EventType = "COMMAND_START"
	EventTypeCommandEnd   EventType = "COMMAND_END"
	EventTypeCommandError EventType = "COMMAND_ERROR"
)

// set of event data keys
const (
	eventDataKeyCommand     = "command"
	eventDataKeyExecutionID = "xid"
	eventDataKeyVersion     = "version"

	EventDataKeyError = "err"
)
