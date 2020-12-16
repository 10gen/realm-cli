package terminal

import (
	"fmt"
	"strings"
)

var (
	listFields = []string{logFieldMessage, logFieldData}
)

type list struct {
	message string
	data    []string
	width   int
}

func newList(message string, data []interface{}) list {
	var l list

	l.message = message
	l.data = make([]string, 0, len(data))
	l.width = 0

	for _, item := range data {
		parsedValue := parseValue(item)
		l.data = append(l.data, parsedValue)
		rowWidth := len(parsedValue)
		if rowWidth > l.width {
			l.width = rowWidth
		}
	}
	return l
}

func (l list) Message() (string, error) {
	return fmt.Sprintf(`%s
%s
%s`, l.message, l.dividerString(), l.dataString()), nil
}

func (l list) Payload() ([]string, map[string]interface{}, error) {
	return listFields, map[string]interface{}{
		logFieldMessage: l.message,
		logFieldData:    l.data,
	}, nil
}

func (l list) dataString() string {
	return strings.Join(l.data, "\n")
}

func (l list) dividerString() string {
	return strings.Repeat("-", l.width)
}
