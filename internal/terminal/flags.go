package terminal

import (
	"fmt"
	"strings"
)

// OutputFormat is the terminal output format
type OutputFormat string

// String returns the output format display
func (of OutputFormat) String() string {
	val := string(of)
	return val
}

// Type returns the OutputFormat type
func (of OutputFormat) Type() string { return "string" }

// Set validates and sets the output format value
func (of *OutputFormat) Set(val string) error {
	outputFormat := OutputFormat(val)

	if !isValidOutputFormat(outputFormat) {
		allOutputFormats := []string{OutputFormatJSON.String()}
		return fmt.Errorf("unsupported value, use one of [%s] instead", strings.Join(allOutputFormats, ", "))
	}

	*of = outputFormat
	return nil
}

// set of supported terminal output formats
const (
	OutputFormatText OutputFormat = "" // zero-valued to be flag's default
	OutputFormatJSON OutputFormat = "json"
)

func isValidOutputFormat(outputFormat OutputFormat) bool {
	switch outputFormat {
	case
		OutputFormatJSON,
		OutputFormatText:
		return true
	}
	return false
}
