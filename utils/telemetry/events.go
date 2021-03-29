package telemetry

import (
	"fmt"
	"strings"
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

type Mode string

const (
	ModeOff   Mode = "off"
	ModeOn    Mode = "on"
	ModeEmpty Mode = ""
)

func isValid(mode Mode) bool {
	switch mode {
	case ModeOn, ModeOff, ModeEmpty:
		return true
	}
	return false
}

func (m Mode) String() string {
	return string(m)
}
func (m *Mode) Set(value string) error {
	mode := Mode(value)
	if !isValid(mode) {
		allModes := []string{string(ModeOn), string(ModeOff)}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allModes, ", "))
	}
	*m = mode
	return nil
}
