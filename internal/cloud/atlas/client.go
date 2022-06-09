package atlas

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/utils/api"

	"github.com/edaniels/digest"
)

const (
	publicAPI = "/api/public/v1.0"
	atlasAPI  = "/api/atlas/v1.0"

	userAgentHeader = "User-Agent"
	cliHeaderValue  = "MongoDB-BaaS-CLI"
)

// Client is a MongoDB Cloud Atlas client
type Client interface {
	Groups(url string, useBaseURL bool) (Groups, error)

	Clusters(groupID string) ([]Cluster, error)
	ServerlessInstances(groupID string) ([]ServerlessInstance, error)
	Datalakes(groupID string) ([]Datalake, error)

	Status() error
}

// NewClient returns a new MongoDB Cloud Atlas client
func NewClient(baseURL string) Client {
	return &client{baseURL: baseURL}
}

// NewAuthClient returns a new authenticated MongoDB Cloud Atlas client
func NewAuthClient(baseURL string, creds user.Credentials) Client {
	return &client{
		baseURL:   baseURL,
		transport: digest.NewTransport(creds.PublicAPIKey, creds.PrivateAPIKey),
	}
}

type client struct {
	baseURL   string
	transport *digest.Transport
}

func (c *client) doWithBaseURL(method, path string, options api.RequestOptions) (*http.Response, error) {
	return c.doWithURL(method, c.baseURL+path, options)
}

func (c *client) doWithURL(method, url string, options api.RequestOptions) (*http.Response, error) {
	req, reqErr := http.NewRequest(method, url, options.Body)
	if reqErr != nil {
		return nil, reqErr
	}

	api.IncludeQuery(req, options.Query)

	req.Header.Set(userAgentHeader, cliHeaderValue)

	if options.ContentType != "" {
		req.Header.Set(api.HeaderContentType, options.ContentType)
	}

	client := &http.Client{}
	client.Timeout = time.Second * 20

	if c.transport == nil {
		if !options.NoAuth {
			return nil, ErrMissingAuth
		}
		return client.Do(req)
	}
	client.Transport = c.transport

	res, resErr := client.Do(req)
	if resErr != nil {
		if netErr, ok := resErr.(net.Error); ok && netErr.Timeout() {
			return nil, errServerError{"request timed out after " + client.Timeout.String()}
		}
		return nil, errServerError{}
	}

	if res.StatusCode == http.StatusUnauthorized {
		defer res.Body.Close()

		var errRes errResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return nil, ErrUnauthorized{err.Error()}
		}
		return nil, ErrUnauthorized{errRes.Detail}
	}

	if res.StatusCode == http.StatusForbidden {
		return nil, errForbidden(res.Status)
	}

	return res, nil
}
