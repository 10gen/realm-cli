package terminal

import "fmt"

const (
	logFieldMessage = "message"
)

var (
	textMessageFields = []string{logFieldMessage}
)

type textMessage string

func newTextMessage(format string, args ...interface{}) textMessage {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	return textMessage(message)
}

func (t textMessage) Message() (string, error) {
	return string(t), nil
}

func (t textMessage) Payload() ([]string, map[string]interface{}, error) {
	return textMessageFields, map[string]interface{}{
		logFieldMessage: t,
	}, nil
}
