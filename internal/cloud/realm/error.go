package realm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

const (
	invalidSessionCode = "InvalidSession"
)

// ErrInvalidSession is an invalid session error that has a follow up message
type ErrInvalidSession struct {
	error
}

func (err ErrInvalidSession) SuggestedCommands() []string {
	return []string{}
}

func (err ErrInvalidSession) ReferenceLinks() []string {
	return []string{}
}

func NewErrInvalidSession() ErrInvalidSession {
	return ErrInvalidSession{errors.New("invalid session") }
}

// ServerError is a Realm server error
type ServerError struct {
	Code    string `json:"error_code"`
	Message string `json:"error"`
}

func (se ServerError) Error() string {
	return se.Message
}

// parseResponseError attempts to read and unmarshal a server error
// from the provided *http.Response
func parseResponseError(res *http.Response) error {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return err
	}

	payload := buf.String()
	if payload == "" {
		return ServerError{Message: res.Status}
	}

	var serverError ServerError
	if err := json.NewDecoder(buf).Decode(&serverError); err != nil {
		serverError.Message = payload
	}
	return serverError
}
