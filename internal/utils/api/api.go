package api

import (
	"fmt"
	"io"
	"net/http"
)

// set of supported api header keys
const (
	HeaderAccept                  = "Accept"
	HeaderCacheControl            = "Cache-Control"
	HeaderContentDisposition      = "Content-Disposition"
	HeaderContentEncoding         = "Content-Encoding"
	HeaderContentLanguage         = "Content-Language"
	HeaderContentType             = "Content-Type"
	HeaderAuthorization           = "Authorization"
	HeaderWebsiteRedirectLocation = "Website-Redirect-Location"
)

// set of supported api media types
const (
	MediaTypeJSON = "application/json"
)

// RequestOptions are options to configure an *http.Request
type RequestOptions struct {
	Body           io.Reader
	ContentType    string
	NoAuth         bool
	PreventRefresh bool
	Query          map[string]string
	RefreshAuth    bool
}

// IncludeQuery includes the query with the http request
func IncludeQuery(req *http.Request, q map[string]string) {
	if len(q) > 0 {
		query := req.URL.Query()
		for k, v := range q {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}
}

// ErrUnexpectedStatusCode is an unexpected status code error
type ErrUnexpectedStatusCode struct {
	Action string
	Actual int
}

func (err ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf(
		"failed to %s: unexpected status code %d",
		err.Action,
		err.Actual,
	)
}
