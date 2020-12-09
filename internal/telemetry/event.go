package telemetry

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event is a telemetry event
type Event struct {
	id          string
	eventType   EventType
	userID      string
	time        time.Time
	executionID string
	command     string
	data        map[DataKey]interface{}
}

// NewCommandStartEvent creates a new event for a command's start
func NewCommandStartEvent(command string) Event {
	data := make(map[DataKey]interface{})
	return NewEvent(EventTypeCommandStart, command, data)
}

// NewCommandCompleteEvent creates a new event for a command's start
func NewCommandCompleteEvent(command string) Event {
	data := make(map[DataKey]interface{})
	return NewEvent(EventTypeCommandStart, command, data)
}

// NewCommandErrorEvent creates a new event for a command that has errored
func NewCommandErrorEvent(command string, err error) Event {
	data := make(map[DataKey]interface{})
	data[DataKeyErr] = err
	return NewEvent(EventTypeCommandStart, command, data)
}

// NewEvent creates a new event
func NewEvent(eventType EventType, command string, data map[DataKey]interface{}) Event {
	return Event{
		id:        primitive.NewObjectID().Hex(),
		eventType: eventType,
		time:      time.Now(),
		command:   command,
		data:      data,
	}
}

// EventType is a cli event type
type EventType string

// set of supported cli event types
const (
	EventTypeCommandStart    EventType = "COMMAND_START"
	EventTypeCommandComplete EventType = "COMMAND_COMPLETE"
	EventTypeCommandError    EventType = "COMMAND_ERROR"
)

// DataKey used to pass data into the Event.Data map
type DataKey string

// set of Data Keys
const (
	DataKeyErr DataKey = "err"
)
