package telemetry

import (
	"time"
)

// REALMC-7243 same as the new cli Segment tracking

type event struct {
	id          string
	eventType   EventType
	userID      string
	time        time.Time
	executionID string
	command     string
	data        []EventData
}

type EventData struct {
	Key   string
	Value interface{}
}

type EventType string

const (
	EventTypeCommandStart EventType = "COMMAND_START"
	EventTypeCommandEnd   EventType = "COMMAND_END"
	EventTypeCommandError EventType = "COMMAND_ERROR"
)

const (
	eventDataKeyCommand     = "command"
	eventDataKeyExecutionID = "xid"

	EventDataKeyError = "err"
)

type Mode bool
