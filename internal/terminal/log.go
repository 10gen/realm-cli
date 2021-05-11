package terminal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/iancoleman/orderedmap"
)

// LogLevel is the level of a terminal log
type LogLevel string

// set of supported log levels
const (
	LogLevelInfo  LogLevel = "info"
	LogLevelError LogLevel = "error"
	LogLevelWarn  LogLevel = "warn"
	LogLevelDebug LogLevel = "debug"
)

// LogData produces the log data
type LogData interface {
	Message() (string, error)
	Payload() ([]string, map[string]interface{}, error)
}

// Log is a terminal log
type Log struct {
	Level LogLevel
	Time  time.Time
	Data  LogData
}

// NewDebugLog creates a new debug log with a text message
func NewDebugLog(format string, args ...interface{}) Log {
	return newLog(LogLevelDebug, newTextMessage(format, args...))
}

// NewTextLog creates a new log with a text message
func NewTextLog(format string, args ...interface{}) Log {
	return newLog(LogLevelInfo, newTextMessage(format, args...))
}

// NewJSONLog creates a new log with a JSON document
func NewJSONLog(message string, data interface{}) Log {
	return newLog(LogLevelInfo, jsonDocument{message, data})
}

// NewTableLog creates a new log with a table
func NewTableLog(message string, headers []string, data ...map[string]interface{}) Log {
	return newLog(LogLevelInfo, newTable(message, headers, data))
}

// NewListLog creates a new log with a list
func NewListLog(message string, data ...interface{}) Log {
	return newLog(LogLevelInfo, newList(message, data, false))
}

// NewErrorLog creates a new error log
func NewErrorLog(err error) Log {
	return newLog(LogLevelError, errorMessage{err})
}

// NewWarningLog creates a new warning log
func NewWarningLog(format string, args ...interface{}) Log {
	return newLog(LogLevelWarn, newTextMessage(format, args...))
}

// NewFollowupLog creates a new log with a consolidated list of followup items
func NewFollowupLog(message string, items ...interface{}) Log {
	return newLog(LogLevelDebug, newList(message, items, true))
}

// set of common log messages
const (
	MsgReferenceLinks string = "For more information"
	MsgSuggestions    string = "Try instead"
)

func newLog(level LogLevel, data LogData) Log {
	return Log{level, time.Now(), data}
}

// Print produces the log output based on the specified format
func (l Log) Print(outputFormat OutputFormat) (string, error) {
	switch outputFormat {
	case OutputFormatText:
		return l.textLog()
	case OutputFormatJSON:
		return l.jsonOutput()
	default:
		return "", fmt.Errorf("unsupported output format type: %s", outputFormat)
	}
}

func (l Log) textLog() (string, error) {
	message, err := l.Data.Message()
	if err != nil {
		return "", err
	}

	return message, nil
}

const (
	logFieldLevel = "level"
	logFieldTime  = "time"
)

func (l Log) jsonOutput() (string, error) {
	out := orderedmap.New()
	out.Set(logFieldTime, l.Time)
	out.Set(logFieldLevel, l.Level)

	keys, payload, err := l.Data.Payload()
	if err != nil {
		return "", err
	}
	for _, key := range keys {
		out.Set(key, payload[key])
	}

	output, outputErr := json.Marshal(out)
	return string(output), outputErr
}
