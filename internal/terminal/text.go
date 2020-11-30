package terminal

const (
	logFieldMessage = "message"
)

var (
	textMessageFields = []string{logFieldMessage}
)

type textMessage string

func (t textMessage) Message() (string, error) {
	return string(t), nil
}

func (t textMessage) Payload() ([]string, map[string]interface{}, error) {
	return textMessageFields, map[string]interface{}{
		logFieldMessage: t,
	}, nil
}
