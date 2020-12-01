package terminal

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

const (
	logFieldTitle = "title"
	logFieldDoc   = "doc"
)

var (
	jsonDocumentFields       = []string{logFieldDoc}
	titledJSONDocumentFields = []string{logFieldTitle, logFieldDoc}
)

type jsonDocument struct {
	data interface{}
}

func (j jsonDocument) Message() (string, error) {
	data, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

func (j jsonDocument) Payload() ([]string, map[string]interface{}, error) {
	return jsonDocumentFields, map[string]interface{}{
		logFieldDoc: j.data,
	}, nil
}

type titledJSONDocument struct {
	title string
	jsonDocument
}

func (tj titledJSONDocument) Message() (string, error) {
	doc, err := tj.jsonDocument.Message()
	if err != nil {
		return "", err
	}

	title := color.New(color.Bold).SprintFunc()(tj.title)
	return fmt.Sprintf("%s\n---\n%s", title, doc), nil
}

func (tj titledJSONDocument) Payload() ([]string, map[string]interface{}, error) {
	return titledJSONDocumentFields, map[string]interface{}{
		logFieldTitle: tj.title,
		logFieldDoc:   tj.jsonDocument.data,
	}, nil
}
