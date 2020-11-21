package realm

import (
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	adminAPI   = "/api/admin/v3.0"
	privateAPI = "/api/private/v1.0"

	requestOriginHeader = "X-BAAS-Request-Origin"
	cliHeaderValue      = "mongodb-baas-cli"
)

// Client is a Realm client
type Client interface {
	Authenticate(publicAPIKey, privateAPIKey string) (AuthResponse, error)
	Status() error
}

// NewClient creates a new Realm client
func NewClient(baseURL string) Client {
	return &client{baseURL}
}

type client struct {
	baseURL string
}

func (c *client) do(method, path string, options api.RequestOptions) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, options.Body)
	if err != nil {
		return nil, err
	}

	req.Header = options.Header
	if req.Header == nil {
		req.Header = http.Header{}
	}
	req.Header.Set(requestOriginHeader, cliHeaderValue)

	client := &http.Client{}
	return client.Do(req)
}
