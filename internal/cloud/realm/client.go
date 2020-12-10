package realm

import (
	"bytes"
	"encoding/json"
	"errors"
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
	AuthProfile() (AuthProfile, error)
	Authenticate(publicAPIKey, privateAPIKey string) (Session, error)

	FindApps(filter AppFilter) ([]App, error)

	CreateAPIKey(groupID, appID, apiKeyName string) (APIKey, error)
	CreateUser(groupID, appID, email, password string) (User, error)

	Status() error
}

// NewClient creates a new Realm client
func NewClient(baseURL string) Client {
	return &client{baseURL: baseURL}
}

// NewAuthClient creates a new Realm client with a session used for Authorization
func NewAuthClient(baseURL string, session Session) Client {
	return &client{baseURL, session}
}

type client struct {
	baseURL string
	session Session
}

func (c *client) doJSON(method, path string, payload interface{}, options api.RequestOptions) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	options.Body = bytes.NewReader(body)
	options.ContentType = api.MediaTypeApplicationJSON

	return c.do(method, path, options)
}

func (c *client) do(method, path string, options api.RequestOptions) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, options.Body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(requestOriginHeader, cliHeaderValue)

	if options.ContentType != "" {
		req.Header.Set(api.HeaderContentType, options.ContentType)
	}

	if auth, err := c.getAuth(options); err != nil {
		return nil, err
	} else if auth != "" {
		req.Header.Set(api.HeaderAuthorization, "Bearer "+auth)
	}
	client := &http.Client{}
	return client.Do(req)
}

func (c *client) getAuth(options api.RequestOptions) (string, error) {
	if options.UseAuth {
		if c.session.AccessToken == "" {
			return "", errors.New("required access token is blank")
		}
		return c.session.AccessToken, nil
	}

	if options.RefreshAuth {
		if c.session.RefreshToken == "" {
			return "", errors.New("required refresh token is blank")
		}
		return c.session.RefreshToken, nil
	}

	return "", nil
}
