package realm

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/10gen/realm-cli/internal/auth"
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
	ExportDependencies(groupID, appID string) (string, io.ReadCloser, error)
	Import(groupID, appID string, appData interface{}) error
	ImportDependencies(groupID, appID, uploadPath string) error
	Diff(groupID, appID string, appData interface{}) ([]string, error)
	DiffDependencies(groupID, appID, uploadPath string) (DependenciesDiff, error)

	CreateApp(groupID, name string, meta AppMeta) (App, error)
	DeleteApp(groupID, appID string) error
	FindApps(filter AppFilter) ([]App, error)
	AppDescription(groupID, appID string) (AppDescription, error)

	CreateDraft(groupID, appID string) (AppDraft, error)
	DeployDraft(groupID, appID, draftID string) (AppDeployment, error)
	DiffDraft(groupID, appID, draftID string) (AppDraftDiff, error)
	DiscardDraft(groupID, appID, draftID string) error
	Deployments(groupID, appID string) ([]AppDeployment, error)
	Deployment(groupID, appID, deploymentID string) (AppDeployment, error)
	Draft(groupID, appID string) (AppDraft, error)

	Secrets(groupID, appID string) ([]Secret, error)
	CreateSecret(groupID, appID, name, value string) (Secret, error)
	DeleteSecret(groupID, appID, secretID string) error
	UpdateSecret(groupID, appID, secretID, name, value string) error

	CreateAPIKey(groupID, appID, apiKeyName string) (APIKey, error)
	CreateUser(groupID, appID, email, password string) (User, error)
	DeleteUser(groupID, appID, userID string) error
	DisableUser(groupID, appID, userID string) error
	EnableUser(groupID, appID, userID string) error
	FindUsers(groupID, appID string, filter UserFilter) ([]User, error)
	RevokeUserSessions(groupID, appID, userID string) error

	HostingAssets(groupID, appID string) ([]HostingAsset, error)
	HostingAssetUpload(groupID, appID, rootDir string, asset HostingAsset) error
	HostingAssetRemove(groupID, appID, path string) error
	HostingAssetAttributesUpdate(groupID, appID, path string, attrs ...HostingAssetAttribute) error
	HostingCacheInvalidate(groupID, appID, path string) error

	Functions(groupID, appID string) ([]Function, error)
	AppDebugExecuteFunction(groupID, appID, userID, name string, args []interface{}) (ExecutionResults, error)

	Logs(groupID, appID string, opts LogsOptions) (Logs, error)

	Status() error
}

// NewClient creates a new Realm client
func NewClient(baseURL string) Client {
	return &client{baseURL, noopAuth{}}
}

// NewAuthClient creates a new Realm client capable of managing the user's session
func NewAuthClient(baseURL string, authService auth.Service) Client {
	return &client{baseURL, authService}
}

type client struct {
	baseURL     string
	authService auth.Service
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

	api.IncludeQuery(req, options.Query)

	req.Header.Set(requestOriginHeader, cliHeaderValue)

	if options.ContentType != "" {
		req.Header.Set(api.HeaderContentType, options.ContentType)
	}

	if token, err := c.getAuthToken(options); err != nil {
		return nil, err
	} else if token != "" {
		req.Header.Set(api.HeaderAuthorization, "Bearer "+token)
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
	} else if options.PreventRefresh || err.Code != errCodeInvalidSession {
		return nil, err
	}

	if refreshErr := c.refreshAuth(); refreshErr != nil {
		c.authService.ClearSession()
		if err := c.authService.Save(); err != nil {
			return nil, ErrInvalidSession{}
		}
		return nil, ErrInvalidSession{}
	}

	options.PreventRefresh = true

	return c.do(method, path, options)
}

type noopAuth struct{}

func (sm noopAuth) ClearSession() {}

func (sm noopAuth) ClearUser() {}

func (sm noopAuth) Save() error { return nil }

func (sm noopAuth) Session() auth.Session { return auth.Session{} }

func (sm noopAuth) SetSession(session auth.Session) {}

func (sm noopAuth) User() auth.User { return auth.User{} }

func (sm noopAuth) SetUser(user auth.User) {}
