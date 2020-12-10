package api

import (
	"io"
)

// set of supported api header keys
const (
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
)

// set of supported api media types
const (
	MediaTypeApplicationJSON = "application/json"
)

// RequestOptions are options to configure an *http.Request
type RequestOptions struct {
	Body        io.Reader
	ContentType string
	UseAuth     bool
	RefreshAuth bool
}
