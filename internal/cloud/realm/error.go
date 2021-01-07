package realm

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	// InvalidSessionCode is the error code returned when the user's sesison is invalid
	invalidSessionCode = "InvalidSession"
)

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
	if res.Header.Get(api.HeaderContentType) != api.MediaTypeJSON {
		return errors.New(res.Status)
	}
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
