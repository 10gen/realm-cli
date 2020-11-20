package terminal

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

// JSONDocument is a JSON document message to display in the UI
type JSONDocument struct {
	Data          interface{}
	NoPrettyPrint bool
}

func (j JSONDocument) marshal() ([]byte, error) {
	if j.NoPrettyPrint {
		return json.Marshal(j.Data)
	}
	return json.MarshalIndent(j.Data, "", "  ")
}

// Message returns a JSON document message, or any error that occurred
// while marshalling the document
func (j JSONDocument) Message() (string, error) {
	data, err := j.marshal()
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

// TitledJSONDocument is a JSON document message with a title to display in the UI
type TitledJSONDocument struct {
	JSONDocument
	Title string
}

// Message returns a titled JSON document menssage, or any error that occurred
// while marshalling the document
func (tj TitledJSONDocument) Message() (string, error) {
	doc, err := tj.JSONDocument.Message()
	if err != nil {
		return "", err
	}

	title := color.New(color.Bold).SprintFunc()(tj.Title)

	return fmt.Sprintf("%s\n---\n%s", title, doc), nil
}
