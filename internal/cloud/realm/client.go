package realm

import (
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
	Authenticate(publicAPIKey, privateAPIKey string) (AuthResponse, error)
	GetUserProfile() (UserProfile, error)
	FindProjectAppByClientAppID(groupIDs []string, app string) ([]App, error)
	Status() error
}

// NewClient creates a new Realm client
func NewClient(baseURL string) Client {
	return &client{baseURL: baseURL}
}

// NewAuthClient creates a new Realm client with a session used for Authorization
func NewAuthClient(baseURL string, session *Session) Client {
	// TODO: REALMC-7156 should we return an error here?
	return &client{baseURL, session}
}

type client struct {
	baseURL string
	session *Session
}

// Session is the CLI profile session TODO REALMC-7156 figure out this approach i.e. moving Session to here
type Session struct {
	AccessToken  string
	RefreshToken string
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

	// TODO REALMC-7156 How do we want to handle refresh tokens here
	if c.session != nil {
		if len(c.session.AccessToken) == 0 {
			return nil, errors.New("the current session is invalid. This can happen if you have not yet logged in or if your refresh token has expired")
		}
		req.Header.Add("Authorization", "Bearer "+c.session.AccessToken)
	}

	client := &http.Client{}
	return client.Do(req)
}
