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

// EventData holds additional event information
type EventData struct {
	Key   string
	Value interface{}
}

// EventType is a cli event type
type EventType string

// set of supported cli event types
const (
	EventTypeCommandStart        EventType = "COMMAND_START"
	EventTypeCommandComplete     EventType = "COMMAND_COMPLETE"
	EventTypeCommandError        EventType = "COMMAND_ERROR"
	EventTypeCommandVersionCheck EventType = "COMMAND_VERSION_CHECK"
)

// set of event data keys
const (
	eventDataKeyCommand     = "cmd"
	eventDataKeyExecutionID = "xid"
	eventDataKeyVersion     = "v"

	EventDataKeyError = "err"
)
