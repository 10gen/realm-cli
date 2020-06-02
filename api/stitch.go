package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrAppNotFound is used when an app cannot be found by client app ID
type ErrAppNotFound struct {
	ClientAppID string
}

func (eanf ErrAppNotFound) Error() string {
	return fmt.Sprintf("Unable to find app with ID: %q", eanf.ClientAppID)
}

// ErrRealmResponse represents a response from a Realm API call
type ErrRealmResponse struct {
	data errRealmResponseData
}

// Error returns a stringified error message
func (esr ErrRealmResponse) Error() string {
	return fmt.Sprintf("error: %s", esr.data.Error)
}

// ErrorCode returns this ErrorCode on the error
func (esr ErrRealmResponse) ErrorCode() string {
	return esr.data.ErrorCode
}

// UnmarshalJSON unmarshals JSON data into an ErrRealmResponse
func (esr *ErrRealmResponse) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &esr.data)
}

type errRealmResponseData struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
}

// UnmarshalRealmError unmarshals an *http.Response into an ErrRealmResponse. If the Body does not
// contain content it uses the provided Status
func UnmarshalRealmError(res *http.Response) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return err
	}

	str := buf.String()
	if str == "" {
		return ErrRealmResponse{
			data: errRealmResponseData{
				Error: res.Status,
			},
		}
	}

	var realmResponse ErrRealmResponse
	if err := json.NewDecoder(&buf).Decode(&realmResponse); err != nil {
		realmResponse.data.Error = str
	}

	return realmResponse
}
