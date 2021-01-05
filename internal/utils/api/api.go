package api

import (
	"io"
)

// set of supported api header keys
const (
	HeaderAccept             = "Accept"
	HeaderContentDisposition = "Content-Disposition"
	HeaderContentType        = "Content-Type"
	HeaderAuthorization      = "Authorization"
)

// set of supported api media types
const (
	MediaTypeJSON = "application/json"
)

// RequestOptions are options to configure an *http.Request
type RequestOptions struct {
	Body           io.Reader
	ContentType    string
	Query          map[string]string
	RefreshAuth    bool
	UseAuth        bool
	PreventRefresh bool
}
