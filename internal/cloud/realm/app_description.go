package realm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	appDescriptionPathPattern = appPathPattern + "/description"
)

// DataSourceSummary is a short summary for a data source model
type DataSourceSummary struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	DataSource string `json:"data_source"`
}

// IncomingWebhookSummary is a short summary for a webhook model
type IncomingWebhookSummary struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// HTTPEndpointSummary is a short summary for a http endpoint model
type HTTPEndpointSummary struct {
	Name             string                   `json:"name"`
	IncomingWebhooks []IncomingWebhookSummary `json:"webhooks"`
}

// ServiceSummary is a short summary for a service desc model
type ServiceSummary struct {
	Name             string                   `json:"name"`
	Type             string                   `json:"type"`
	IncomingWebhooks []IncomingWebhookSummary `json:"webhooks"`
}

// AuthProviderSummary is a short summary for a auth provider config model
type AuthProviderSummary struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// CustomUserDataSummary is a short summary for a custom user data config model
type CustomUserDataSummary struct {
	Enabled     bool   `json:"enabled"`
	DataSource  string `json:"data_source"`
	Database    string `json:"database"`
	Collection  string `json:"collection"`
	UserIDField string `json:"user_id_field"`
}

// HostingSummary is a short summary for a hosting model
type HostingSummary struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	URL     string `json:"url"`
}

// FunctionSummary is a short summary for a function model
type FunctionSummary struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// SyncSummary is a short summary for a sync model
type SyncSummary struct {
	State                  string `json:"state"`
	DataSource             string `json:"data_source"`
	Database               string `json:"database"`
	DevelopmentModeEnabled bool   `json:"development_mode_enabled"`
}

// GraphQLSummary is a short summary for a graphql model
type GraphQLSummary struct {
	URL             string   `json:"url"`
	CustomResolvers []string `json:"custom_resolvers"`
}

// EventSubscriptionSummary is a short summary for a event subscription model
type EventSubscriptionSummary struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type LogForwarderSummary struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// AppDescription describes an App
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
	LogForwarder      []LogForwarderSummary      `json:"log_forwarder"`
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
