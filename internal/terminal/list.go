package terminal

import (
	"fmt"
	"strings"
)

var (
	listFields = []string{logFieldMessage, logFieldData}
)

type list struct {
	message      string
	data         []string
	consolidated bool
}

func newList(message string, data []interface{}, consolidated bool) list {
	l := list{
		message:      message,
		data:         make([]string, 0, len(data)),
		consolidated: consolidated,
	}
	for _, item := range data {
		l.data = append(l.data, parseValue(item))
	}
	return l
}

func (l list) Message() (string, error) {
	if len(l.data) == 0 {
		return l.message, nil
	}

	if l.consolidated && len(l.data) == 1 {
		if len(l.message) == 0 {
			return l.data[0], nil
		}
		return fmt.Sprintf("%s: %s", l.message, l.data[0]), nil
	}

	return fmt.Sprintf("%s\n%s", l.message, l.dataString()), nil
}

func (l list) Payload() ([]string, map[string]interface{}, error) {
	return listFields, map[string]interface{}{
		logFieldMessage: l.message,
		logFieldData:    l.data,
	}, nil
}

func (l list) dataString() string {
	data := make([]string, 0, len(l.data))
	for _, item := range l.data {
		data = append(data, Indent+item)
	}
	return strings.Join(data, "\n")
}
