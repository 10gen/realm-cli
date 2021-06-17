package telemetry

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/flags"
)

// set of supported telemetry flags
const (
	FlagMode      = "telemetry"
	FlagModeUsage = `Enable/Disable CLI usage tracking for your current profile (Default value: "on"; Allowed values: "on", "off")`
)

// Mode is the Telemetry Mode
type Mode string

// String returns the string representation
func (m Mode) String() string { return string(m) }

// Type returns the Mode type
func (m Mode) Type() string { return flags.TypeString }

// Set validates and sets the mode value
func (m *Mode) Set(val string) error {
	mode := Mode(val)

	if !isValidMode(mode) {
		allModes := []string{string(ModeOn), string(ModeOff)}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allModes, ", "))
	}

	*m = mode
	return nil
}

// set of supported telemetry modes
const (
	ModeEmpty  Mode = "" // zero-valued to be flag's default
	ModeOn     Mode = "on"
	ModeStdout Mode = "stdout"
	ModeOff    Mode = "off"
)

func isValidMode(mode Mode) bool {
	switch mode {
	case
		ModeOn,
		ModeEmpty,
		ModeStdout,
		ModeOff:
		return true
	}
	return false
}
