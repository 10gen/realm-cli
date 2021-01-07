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

// NewTextLog creates a new log with a text message
func NewTextLog(format string, args ...interface{}) Log {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	return newLog(LogLevelInfo, textMessage(message))
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
	return newLog(LogLevelInfo, newList(message, data))
}

// NewErrorLog creates a new error log
func NewErrorLog(err error) Log {
	return newLog(LogLevelError, errorMessage{err})
}

// NewWarningLog creates a new warn log
func NewWarningLog(format string, args ...interface{}) Log {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	return newLog(LogLevelInfo, textMessage(message))
}

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
