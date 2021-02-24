package terminal

import (
	"encoding/json"
	"fmt"
	"strings"
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

var (
	allLogLevels = []LogLevel{LogLevelInfo, LogLevelError}

	longestLogLevel = func() LogLevel {
		var longest LogLevel
		for _, level := range allLogLevels {
			if len(level) > len(longest) {
				longest = level
			}
		}
		return longest
	}()
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

// NewFollowupLog creates a log with a follow up message about suggested commands for the error
func NewFollowupLog(message string, items ...interface{}) Log {
	return newLog(LogLevelDebug, newList(message, items, true))
}

// Default display verbage
const (
	LinkMessage    string = "For more information"
	CommandMessage string = "Try running instead"
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

	return fmt.Sprintf(
		fmt.Sprintf("%%s UTC %%-%ds %%s", len(longestLogLevel)),
		l.Time.In(time.UTC).Format("15:04:05"),
		strings.ToUpper(string(l.Level)),
		message,
	), nil
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
