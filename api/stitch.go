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

// ErrStitchResponse represents a response from a Stitch API call
type ErrStitchResponse struct {
	data errStitchResponseData
}

// Error returns a stringified error message
func (esr ErrStitchResponse) Error() string {
	return fmt.Sprintf("error: %s", esr.data.Error)
}

// UnmarshalJSON unmarshals JSON data into an ErrStitchResponse
func (esr *ErrStitchResponse) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &esr.data)
}

type errStitchResponseData struct {
	Error string `json:"error"`
}

// UnmarshalStitchError unmarshals an *http.Response into an ErrStitchResponse. If the Body does not
// contain content it uses the provided Status
func UnmarshalStitchError(res *http.Response) error {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return err
	}

	str := buf.String()
	if str == "" {
		return ErrStitchResponse{
			data: errStitchResponseData{
				Error: res.Status,
			},
		}
	}

	var stitchResponse ErrStitchResponse
	if err := json.NewDecoder(&buf).Decode(&stitchResponse); err != nil {
		stitchResponse.data.Error = str
	}

	return stitchResponse
}
