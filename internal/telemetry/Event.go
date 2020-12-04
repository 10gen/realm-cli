package telemetry

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event is a cli event
type Event struct {
	ID     primitive.ObjectID
	Type   EventType
	UserID string
	Time   time.Time
	Data   map[DataKey]interface{}
}

// EventType is a cli event type
type EventType string

// set of supported cli event types
const (
	EventTypeCommandStart              EventType = "CommandStart"
	EventTypeCommandFinishSuccessfully EventType = "CommandFinishSuccessfully"
	EventTypeCommandError              EventType = "CommandError"
)

// DataKey used to pass data into the Event.Data map
type DataKey string

// set of Data Keys
const (
	CommandKey     DataKey = "CommandKey"
	ExecutionIDKey DataKey = "ExecutionIDKey"
	ErrKey         DataKey = "ErrKey"
)
