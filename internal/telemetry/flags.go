package telemetry

import (
	"fmt"
	"strings"
)

// Mode is the Telemetry Mode
type Mode string

// String returns the string representation
func (m Mode) String() string {
	val := string(m)
	return val
}

// Type returns the Mode type
func (m Mode) Type() string { return "string" }

// Set validates and sets the output format value
func (m *Mode) Set(val string) error {
	mode := Mode(val)

	if !isValidMode(mode) {
		allModes := []string{OnSelected.String(),
			STDOut.String(),
			Off.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allModes, ", "))
	}

	*m = mode
	return nil
}

// set of supported telemetry modes
const (
	//User does not select an option
	OnDefault Mode = "" // zero-valued to be flag's default
	//User deliberately selects this option
	OnSelected Mode = "on"
	//User selects stdout
	STDOut Mode = "stdout"
	//User disables tracking
	Off Mode = "off"
)

func isValidMode(mode Mode) bool {
	switch mode {
	case
		OnSelected,
		OnDefault,
		STDOut,
		Off:
		return true
	}
	return false
}
