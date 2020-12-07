package realm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// App represents basic Realm App data
type App struct {
	ID          string `json:"_id"`
	GroupID     string `json:"group_id"`
	ClientAppID string `json:"client_app_id"`
	Name        string `json:"name"`
}

var (
	errGroupNotFound = errors.New("group could not be found")
)

var (
	appsByGroupIDEndpoint = adminAPI + "/groups/%s/apps"
	userProfileEndpoint   = adminAPI + "/auth/profile"
)

func (c *client) GetAppsForUser() ([]App, error) {
	profile, err := c.GetUserProfile()
	if err != nil {
		return nil, err
	}
	groupIDs := profile.AllGroupIDs()
	var appArr []App
	appArr = make([]App, 0)
	for _, groupID := range groupIDs {
		if groupID == "" {
			continue
		}
		apps, err := c.GetApps(groupID)
		if err != nil {
			// Request will fail if any GetApps call fails
			return nil, err
		}
		appSet := make(map[string]bool)
		for _, appElem := range apps {
			if !appSet[appElem.Name] {
				appArr = append(appArr, appElem)
				appSet[appElem.Name] = true
			}
		}
	}
	return appArr, nil
}

// GetApps fetches all Realm Apps associated with the given groupID
func (c *client) GetApps(groupID string) ([]App, error) {
	res, err := c.do(http.MethodGet, fmt.Sprintf(appsByGroupIDEndpoint, groupID), api.RequestOptions{UseAuth: true})

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, errGroupNotFound
		}
		return nil, fmt.Errorf("HTTP status error: %v", res.StatusCode)
	}

	dec := json.NewDecoder(res.Body)
	var apps []App
	if err := dec.Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}
