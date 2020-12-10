package realm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/10gen/realm-cli/internal/utils/api"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	appsPathPattern = adminAPI + "/groups/%s/apps"
)

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
	res, resErr := c.do(http.MethodGet, fmt.Sprintf(appsPathPattern, groupID), api.RequestOptions{UseAuth: true})
	if resErr != nil {
		return nil, resErr
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("group could not be found")
	}
	if res.StatusCode != http.StatusOK {
		return nil, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var apps []App
	if err := dec.Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}

// AppFilter represents the optional filter parameters available for lists of apps
type AppFilter struct {
	GroupID string
	App     string // can be client app id or name
}

// App is a Realm application
type App struct {
	ID              primitive.ObjectID `json:"_id"`
	ClientAppID     string             `json:"client_app_id"`
	Name            string             `json:"name"`
	DomainID        primitive.ObjectID `json:"domain_id"`
	GroupID         string             `json:"group_id"`
	Location        string             `json:"location,omitempty"`
	DeploymentModel string             `json:"deployment_model,omitempty"`
	LastUsed        int64              `json:"last_used"`
	LastModified    int64              `json:"last_modified"`
	Product         string             `json:"product"`
}

func (app App) String() string {
	return fmt.Sprintf("%s (%s)", app.ClientAppID, app.GroupID)
}
