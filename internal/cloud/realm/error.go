package realm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/10gen/realm-cli/internal/cli/feedback"
	"github.com/10gen/realm-cli/internal/cli/user"
)

// set of known error codes
const (
	errCodeInvalidSession = "InvalidSession"

	ErrCodeDraftAlreadyExists = "DraftAlreadyExists"
)

// set of known Realm errors
var (
	ErrDraftNotFound = errors.New("failed to find draft")
)

// ErrInvalidSession is an invalid session error
func ErrInvalidSession(profileName string) error {
	suggestion := "realm-cli login"
	if profileName != user.DefaultProfile {
		suggestion += " --profile " + profileName
	}

	return feedback.NewErr(errors.New("invalid session"), feedback.ErrSuggestion{suggestion})
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
