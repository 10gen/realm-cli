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

// HTTPServiceSummary is a short summary for a service webhook (http endpoint)
type HTTPServiceSummary struct {
	Name             string                   `json:"name"`
	IncomingWebhooks []IncomingWebhookSummary `json:"webhooks"`
}

// EndpointSummary is a short summary for an endpoint
type EndpointSummary struct {
	Route      string `json:"route"`
	HTTPMethod string `json:"http_method"`
	URL        string `json:"url"`
}

type httpEndpointSummary struct {
	HTTPServiceSummary
	EndpointSummary
}

// HTTPEndpoints contains summaries for an http service or endpoint
type HTTPEndpoints struct {
	Summaries []interface{}
}

// MarshalJSON marshals the HTTPEndpoints data to JSON
func (h HTTPEndpoints) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Summaries)
}

// UnmarshalJSON unmarshals JSON into HTTPEndpoints
func (h *HTTPEndpoints) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}

	h.Summaries = make([]interface{}, 0, len(arr))
	for _, m := range arr {
		var t httpEndpointSummary
		if err := json.Unmarshal(m, &t); err != nil {
			return err
		}
		var summary interface{}
		switch {
		case t.Name != "":
			summary = HTTPServiceSummary{t.Name, t.IncomingWebhooks}
		case t.Route != "":
			summary = EndpointSummary{t.Route, t.HTTPMethod, t.URL}
		default:
			summary = t
		}
		h.Summaries = append(h.Summaries, summary)
	}
	return nil
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

// LogForwarderSummary is a short summary for a log forwarder model
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
	HTTPEndpoints     HTTPEndpoints              `json:"http_endpoints"`
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
	LogForwarders     []LogForwarderSummary      `json:"log_forwarders"`
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
