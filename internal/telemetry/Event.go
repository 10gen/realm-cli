package telemetry

import (
	"time"
)

// Event is a cli event
type Event struct {
	ID     string
	Type   EventType
	UserID string    //public api key
	Time   time.Time //uint64
	Data   map[DataKey]interface{}
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
	DataKeyCommand     DataKey = "command"
	DataKeyExecutionID DataKey = "execution_id" //new object id
	DataKeyErr         DataKey = "err"
	//DataKey
)
