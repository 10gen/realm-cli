package telemetry

import (
	"fmt"
	"strings"
)

// Mode is the Telemetry Mode
type Mode string

// String returns the string representation
func (m Mode) String() string { return string(m) }

// Type returns the Mode type
func (m Mode) Type() string { return "string" }

// Set validates and sets the output format value
func (m *Mode) Set(val string) error {
	mode := Mode(val)

	if !isValidMode(mode) {
		allModes := []string{string(ModeOn), string(ModeStdout), string(ModeOff)}
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
	modeTest   Mode = "test" // note: not valid as a command line flag
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
