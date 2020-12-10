package terminal

import (
	"encoding/json"
	"fmt"
)

const (
	logFieldDoc = "doc"
)

var (
	jsonDocumentFields = []string{logFieldMessage, logFieldDoc}
)

type jsonDocument struct {
	message string
	data    interface{}
}

func (j jsonDocument) Message() (string, error) {
	data, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s\n%s", j.message, data), nil
}

func (j jsonDocument) Payload() ([]string, map[string]interface{}, error) {
	return jsonDocumentFields, map[string]interface{}{
		logFieldMessage: j.message,
		logFieldDoc:     j.data,
	}, nil
}
