package realm

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/profile"
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
	Authenticate(publicAPIKey, privateAPIKey string) (auth.Session, error)

	Export(groupID, appID string, req ExportRequest) (string, *zip.Reader, error)
	Import(groupID, appID string, req ImportRequest) error

	CreateApp(groupID, name string, meta AppMeta) (App, error)
	DeleteApp(groupID, appID string) error
	FindApps(filter AppFilter) ([]App, error)

	CreateAPIKey(groupID, appID, apiKeyName string) (APIKey, error)
	CreateUser(groupID, appID, email, password string) (User, error)
	DeleteUser(groupID, appID, userID string) error
	DisableUser(groupID, appID, userID string) error
	FindUsers(groupID, appID string, filter UserFilter) ([]User, error)
	RevokeUserSessions(groupID, appID, userID string) error

	Status() error
}

// NewClient creates a new Realm client
func NewClient(baseURL string) Client {
	return &client{baseURL: baseURL}
}

// NewAuthClient creates a new Realm client with a session used for Authorization
func NewAuthClient(baseURL string, profile *profile.Profile) Client {
	return &client{baseURL, profile}
}

type client struct {
	baseURL string
	profile *profile.Profile
}

func (c *client) doJSON(method, path string, payload interface{}, options api.RequestOptions) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	options.Body = bytes.NewReader(body)
	options.ContentType = api.MediaTypeJSON

	return c.do(method, path, options)
}

func (c *client) do(method, path string, options api.RequestOptions) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, options.Body)
	if err != nil {
		return nil, err
	}

	if len(options.Query) > 0 {
		query := req.URL.Query()
		for key, value := range options.Query {
			query.Add(key, value)
		}
		req.URL.RawQuery = query.Encode()
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

	res, err := client.Do(req)
	if res.StatusCode != http.StatusUnauthorized || !options.UseAuth {
		return res, err
	}

	authToken, refreshErr := c.refreshAuth(client)
	if refreshErr != nil {
		return nil, refreshErr
	}
	req.Header.Set(api.HeaderAuthorization, "Bearer "+authToken)
	// TODO REALMC-7580 change these lines to just return client.Do(req) as session will be cleared downstream
	res, err = client.Do(req)
	if res.StatusCode == http.StatusUnauthorized {
		c.profile.ClearSession()
		return nil, err
	}
	return res, err
}

func (c *client) refreshAuth(httpClient *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+refreshPath, nil)
	if err != nil {
		return "", err
	}

	refreshToken, err := c.getAuth(api.RequestOptions{RefreshAuth: true})
	if err != nil {
		return "", err
	}

	req.Header.Set(api.HeaderAuthorization, "Bearer "+refreshToken)

	res, err := httpClient.Do(req)
	if res.StatusCode == http.StatusUnauthorized {
		return "", ErrInvalidSession
	}
	if res.StatusCode != http.StatusCreated {
		return "", err
	}
	defer res.Body.Close()

	var session auth.Session
	if err := json.NewDecoder(res.Body).Decode(&session); err != nil {
		return "", err
	}
	c.profile.SetSession(session.AccessToken, refreshToken)
	return session.AccessToken, c.profile.Save()
}
