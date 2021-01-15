package realm

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/cli"
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
	return &client{baseURL, noopSessionManager{}}
}

// NewAuthClient creates a new Realm client with a session used for Authorization
func NewAuthClient(profile *cli.Profile) Client {
	return &client{profile.RealmBaseURL(), profile}
}

type client struct {
	baseURL        string
	sessionManager auth.SessionManager
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

	res, resErr := client.Do(req)
	if resErr != nil {
		return nil, resErr
	}

	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return res, nil
	}
	defer res.Body.Close()

	parsedErr := parseResponseError(res)
	if err, ok := parsedErr.(ServerError); !ok {
		return nil, parsedErr
	} else if options.PreventRefresh || err.Code != invalidSessionCode {
		return nil, err
	}

	if refreshErr := c.refreshAuth(); refreshErr != nil {
		c.sessionManager.ClearSession()
		if err := c.sessionManager.Save(); err != nil {
			return nil, ErrInvalidSession{}
		}
		return nil, ErrInvalidSession{}
	}

	options.PreventRefresh = true

	return c.do(method, path, options)
}

type noopSessionManager struct{}

func (sm noopSessionManager) ClearSession() {}

func (sm noopSessionManager) Save() error { return nil }

func (sm noopSessionManager) Session() auth.Session { return auth.Session{} }

func (sm noopSessionManager) SetSession(session auth.Session) {}
