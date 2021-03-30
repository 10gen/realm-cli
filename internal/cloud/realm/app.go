package realm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	appsPathPattern           = adminAPI + "/groups/%s/apps"
	appPathPattern            = appsPathPattern + "/%s"
	appDescriptionPathPattern = appPathPattern + "/description"
)

// DataSourceSummary is a short description for an API data source model
type DataSourceSummary struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	DataSource string `json:"data_source"`
}

// HTTPEndpointSummary is a short description for an API http endpoint model
type HTTPEndpointSummary struct {
	Name string `json:"name"`
}

// ServiceSummary is a short description for an API service desc model
type ServiceSummary struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// AuthProviderSummary is a short description for an API auth provider config model
type AuthProviderSummary struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// CustomUserDataSummary is a short description for an API custom user data config model
type CustomUserDataSummary struct {
	Enabled     bool   `json:"enabled"`
	DataSource  string `json:"data_source"`
	Database    string `json:"database"`
	Collection  string `json:"collection"`
	UserIDField string `json:"user_id_field"`
}

// HostingSummary is a short description for an API hosting model
type HostingSummary struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	URL     string `json:"url"`
}

// FunctionSummary is a short description for an API function model
type FunctionSummary struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// SyncSummary is a short description for an API sync model
type SyncSummary struct {
	State                  string `json:"state"`
	DataSource             string `json:"data_source"`
	Database               string `json:"database"`
	DevelopmentModeEnabled bool   `json:"development_mode_enabled"`
}

// GraphQLSummary is a short description for an API graphql model
type GraphQLSummary struct {
	URL             string   `json:"url"`
	CustomResolvers []string `json:"custom_resolvers"`
}

// EventSubscriptionSummary is a short description for an API event subscription model
type EventSubscriptionSummary struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// AppDescription describes an API App
type AppDescription struct {
	ClientAppID       string                     `json:"client_app_id"`
	Name              string                     `json:"name"`
	RealmURL          string                     `json:"realm_url"`
	DataSources       []DataSourceSummary        `json:"data_sources"`
	HTTPEndpoints     []HTTPEndpointSummary      `json:"http_endpoints"`
	ServiceDescs      []ServiceSummary           `json:"services"`
	AuthProviders     []AuthProviderSummary      `json:"auth_providers"`
	CustomUserData    CustomUserDataSummary      `json:"custom_user_data"`
	Values            []string                   `json:"values"`
	Hosting           HostingSummary             `json:"hosting"`
	Functions         []FunctionSummary          `json:"functions"`
	Sync              SyncSummary                `json:"sync"`
	GraphQL           GraphQLSummary             `json:"graphql"`
	Environment       string                     `json:"environment"`
	EventSubscription []EventSubscriptionSummary `json:"event_subscription"`
}

// AppMeta is Realm application metadata
type AppMeta struct {
	Location        Location        `json:"location,omitempty"`
	DeploymentModel DeploymentModel `json:"deployment_model,omitempty"`
}

// App is a Realm application
type App struct {
	AppMeta
	ID           string `json:"_id"`
	ClientAppID  string `json:"client_app_id"`
	Name         string `json:"name"`
	DomainID     string `json:"domain_id"`
	GroupID      string `json:"group_id"`
	LastUsed     int64  `json:"last_used"`
	LastModified int64  `json:"last_modified"`
	Product      string `json:"product"`
}

// Option returns the Realm app data displayed as a selectable option
func (app App) Option() string {
	return fmt.Sprintf("%s (%s)", app.ClientAppID, app.GroupID)
}

type createAppRequest struct {
	Name string `json:"name"`
	AppMeta
}

func (c *client) CreateApp(groupID, name string, meta AppMeta) (App, error) {
	res, resErr := c.doJSON(
		http.MethodPost,
		fmt.Sprintf(appsPathPattern, groupID),
		createAppRequest{name, meta},
		api.RequestOptions{},
	)
	if resErr != nil {
		return App{}, resErr
	}
	if res.StatusCode != http.StatusCreated {
		return App{}, api.ErrUnexpectedStatusCode{"create app", res.StatusCode}
	}
	defer res.Body.Close()

	var app App
	if err := json.NewDecoder(res.Body).Decode(&app); err != nil {
		return App{}, err
	}
	return app, nil
}

func (c *client) DeleteApp(groupID, appID string) error {
	res, resErr := c.do(
		http.MethodDelete,
		fmt.Sprintf(appPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return resErr
	}
	if res.StatusCode != http.StatusNoContent {
		return api.ErrUnexpectedStatusCode{"delete app", res.StatusCode}
	}
	return nil
}

// AppFilter represents the optional filter parameters available for lists of apps
type AppFilter struct {
	GroupID string
	App     string // can be client app id or name
}

func (c *client) FindApps(filter AppFilter) ([]App, error) {
	var apps []App
	if filter.GroupID == "" {
		arr, err := c.getAppsForUser()
		if err != nil {
			return nil, err
		}
		apps = arr
	} else {
		arr, err := c.getApps(filter.GroupID)
		if err != nil {
			return nil, err
		}
		apps = arr
	}

	if filter.App == "" {
		return apps, nil
	}

	var filtered = make([]App, 0, len(apps))
	for _, app := range apps {
		if strings.HasPrefix(app.ClientAppID, strings.ToLower(filter.App)) {
			filtered = append(filtered, app)
		}
	}
	return filtered, nil
}

func (c *client) getAppsForUser() ([]App, error) {
	profile, profileErr := c.AuthProfile()
	if profileErr != nil {
		return nil, profileErr
	}

	var apps []App
	for _, groupID := range profile.AllGroupIDs() {
		projectApps, err := c.getApps(groupID)
		if err != nil {
			return nil, err
		}
		apps = append(apps, projectApps...)
	}
	return apps, nil
}

func (c *client) getApps(groupID string) ([]App, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(appsPathPattern, groupID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("group could not be found")
	}
	if res.StatusCode != http.StatusOK {
		return nil, api.ErrUnexpectedStatusCode{"get apps", res.StatusCode}
	}
	defer res.Body.Close()

	var apps []App
	if err := json.NewDecoder(res.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *client) AppDescription(groupID, appID string) (AppDescription, error) {
	res, resErr := c.do(
		http.MethodGet,
		fmt.Sprintf(appDescriptionPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if resErr != nil {
		return AppDescription{}, resErr
	}
	if res.StatusCode != http.StatusOK {
		return AppDescription{}, api.ErrUnexpectedStatusCode{"get app description", res.StatusCode}
	}
	var description AppDescription
	if err := json.NewDecoder(res.Body).Decode(&description); err != nil {
		return AppDescription{}, err
	}
	return description, nil
}
