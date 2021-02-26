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
	data        []EventData
}

// EventData holds additional event information
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

// EventDataKey is an event data key
type EventDataKey string

// set of event data keys
const (
	EventDataKeyErr EventDataKey = "err"
)

func eventDataProperties(data []EventData) map[string]interface{} {
	properties := make(map[string]interface{}, len(data))
	for _, datum := range data {
		properties[string(datum.Key)] = datum.Value
	}
	return properties
}
