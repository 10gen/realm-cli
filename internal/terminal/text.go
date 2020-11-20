package terminal

// TextMessage is a text message to display in the UI
type TextMessage struct {
	Contents string
}

// Message returns a text message
func (t TextMessage) Message() (string, error) {
	return t.Contents, nil
}
