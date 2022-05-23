package realm

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/api"
)

const (
	appsPathPattern = adminAPI + "/groups/%s/apps"
	appPathPattern  = appsPathPattern + "/%s"
)

// AppMeta is Realm application metadata
type AppMeta struct {
	Location        Location        `json:"location,omitempty"`
	DeploymentModel DeploymentModel `json:"deployment_model,omitempty"`
	Environment     Environment     `json:"environment,omitempty"`
	Template        string          `json:"template_id,omitempty"`
	DataSource      interface{}     `json:"data_source,omitempty"`
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
	TemplateID   string `json:"template_id"`
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

// TODO(REALMC-9462): remove this once /apps has "template_id" in the payload
func (c *client) FindApp(groupID, appID string) (App, error) {
	res, err := c.do(
		http.MethodGet,
		fmt.Sprintf(appPathPattern, groupID, appID),
		api.RequestOptions{},
	)
	if err != nil {
		return App{}, err
	}
	if res.StatusCode != http.StatusOK {
		return App{}, api.ErrUnexpectedStatusCode{"get app", res.StatusCode}
	}
	defer res.Body.Close()

	var app App
	if err := json.NewDecoder(res.Body).Decode(&app); err != nil {
		return App{}, err
	}
	return app, nil
}

// AppFilter represents the optional filter parameters available for lists of apps
type AppFilter struct {
	GroupID  string
	App      string // can be client app id or name
	Products []string
}

const (
	productStandard = "standard"
	productAtlas    = "atlas"
	productDataAPI  = "data-api"
)

var (
	defaultProducts = []string{productStandard, productAtlas, productDataAPI}
)

func (c *client) FindApps(filter AppFilter) ([]App, error) {
	var apps []App
	if filter.GroupID == "" {
		arr, err := c.getAppsForUser(filter.Products)
		if err != nil {
			return nil, err
		}
		apps = arr
	} else {
		arr, err := c.getApps(filter.GroupID, filter.Products)
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

func (c *client) getAppsForUser(products []string) ([]App, error) {
	profile, profileErr := c.AuthProfile()
	if profileErr != nil {
		return nil, profileErr
	}

	var apps []App
	for _, groupID := range profile.AllGroupIDs() {
		projectApps, err := c.getApps(groupID, products)
		if err != nil {
			return nil, err
		}
		apps = append(apps, projectApps...)
	}
	return apps, nil
}

func (c *client) getApps(groupID string, products []string) ([]App, error) {
	allProducts := resolveProducts(products)

	var apps []App
	for _, product := range allProducts {
		productApps, err := c.getAppsForProduct(groupID, product)
		if err != nil {
			return nil, err
		}
		apps = append(apps, productApps...)
	}
	return apps, nil
}

func resolveProducts(products []string) []string {
	allCap := math.Max(float64(len(defaultProducts)), float64(len(products)))

	allProducts := make([]string, 0, int(allCap))
	allProducts = append(allProducts, products...)

	if len(allProducts) == 0 {
		allProducts = append(allProducts, defaultProducts...)
	}

	return allProducts
}

func (c *client) getAppsForProduct(groupID, product string) ([]App, error) {
	// TODO(REALMC-8886): add tests to verify correct url is being hit
	url := fmt.Sprintf(appsPathPattern, groupID)
	switch product {
	case "", productStandard: // default/empty is considered standard
	default:
		url += "?product=" + product
	}

	res, err := c.do(http.MethodGet, url, api.RequestOptions{})
	if err != nil {
		return nil, err
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
