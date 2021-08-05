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
	eventDataKeyCommand      = "cmd"
	eventDataKeyExecutionID  = "xid"
	eventDataKeyVersion      = "v"
	eventDataKeyError        = "err"
	eventDataKeyErrorMessage = "err_msg"

	// EventDataKeyTemplate used to tracked if templates were used to create an app
	EventDataKeyTemplate = "templateId"
)

// EventDataError returns telemetry event data for an error
func EventDataError(err error) []EventData {
	if err == nil {
		return nil
	}

	return []EventData{
		{eventDataKeyError, err},
		{eventDataKeyErrorMessage, err.Error()},
	}
}

// AdditionalTracker is used to propagate any additional fields to be tracked using our tracking srivce
// this is called AFTER inputs are resolved so it is safe to use any inputs into the commands
type AdditionalTracker interface {
	AdditionalTrackedFields() []EventData
}
