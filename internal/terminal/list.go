package terminal

import (
	"fmt"
	"strings"
)

const (
	// These are messages for the follow ups
	commandMessage = "Try the following command(s)"
	linkMessage    = "Refer to the following link(s)"
)

var (
	listFields = []string{logFieldMessage, logFieldData}
)

type list struct {
	message string
	data    []string
}

func newList(message string, data []interface{}) list {
	l := list{
		message: message,
		data:    make([]string, 0, len(data)),
	}
	for _, item := range data {
		l.data = append(l.data, parseValue(item))
	}
	return l
}

func (l list) Message() (string, error) {
	return fmt.Sprintf("%s%s", l.message, l.dataString()), nil
}

func (l list) Payload() ([]string, map[string]interface{}, error) {
	return listFields, map[string]interface{}{
		logFieldMessage: l.message,
		logFieldData:    l.data,
	}, nil
}

func (l list) dataString() string {
	if len(l.data) == 1 {
		return " " + l.data[0]
	}
	data := make([]string, 0, len(l.data))
	for _, item := range l.data {
		data = append(data, indent+item)
	}
	return "\n" + strings.Join(data, "\n")
}
