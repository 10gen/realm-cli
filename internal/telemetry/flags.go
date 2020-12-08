package telemetry

import (
	"fmt"
	"strings"
)

// Mode is the Telemetry Mode
type Mode string

// NewMode creates a new Mode from the modeString or returns ModeNil
func NewMode(modeString string) Mode {
	mode := Mode(modeString)
	if !isValidMode(mode) {
		return ModeNil
	}
	return mode
}

// String returns the string representation
func (m Mode) String() string { return string(m) }

// Type returns the Mode type
func (m Mode) Type() string { return "string" }

// Set validates and sets the output format value
func (m *Mode) Set(val string) error {
	mode := Mode(val)

	if !isValidMode(mode) {
		allModes := []string{ModeOn.String(), ModeStdout.String(), ModeOff.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allModes, ", "))
	}

	*m = mode
	return nil
}

// set of supported telemetry modes
const (
	ModeNil    Mode = "" // zero-valued to be flag's default
	ModeOn     Mode = "on"
	ModeStdout Mode = "stdout"
	ModeOff    Mode = "off"
)

func isValidMode(mode Mode) bool {
	switch mode {
	case
		ModeOn,
		ModeNil,
		ModeStdout,
		ModeOff:
		return true
	}
	return false
}
