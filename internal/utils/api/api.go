package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// set of supported api header keys
const (
	HeaderContentType = "Content-Type"
)

// set of supported api media types
const (
	MediaTypeApplicationJSON = "application/json"
)

// RequestOptions are options to configure an *http.Request
type RequestOptions struct {
	Body   io.Reader
	Header http.Header
}

// JSONRequestOptions returns RequestOptions configured to send the provided payload as JSON
func JSONRequestOptions(payload interface{}) (RequestOptions, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return RequestOptions{}, err
	}
	return RequestOptions{
		Body:   bytes.NewReader(body),
		Header: http.Header{HeaderContentType: []string{MediaTypeApplicationJSON}},
	}, nil
}
