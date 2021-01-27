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
	appsPathPattern = adminAPI + "/groups/%s/apps"
	appPathPattern  = appsPathPattern + "/%s"
)

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

func (app App) String() string {
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
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return App{}, parseResponseError(res)
	}

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
		defer res.Body.Close()
		return parseResponseError(res)
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
		if strings.HasPrefix(app.ClientAppID, filter.App) {
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
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, parseResponseError(res)
	}

	var apps []App
	if err := json.NewDecoder(res.Body).Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}
