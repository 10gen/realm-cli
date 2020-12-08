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
	appsPathPattern = adminAPI + "/groups/%s/apps"
)

func (c *client) GetAppsForUser() ([]App, error) {
	profile, err := c.GetAuthProfile()
	if err != nil {
		return nil, err
	}
	groupIDs := profile.AllGroupIDs()
	var appArr []App
	appArr = make([]App, 0)
	for _, groupID := range groupIDs {
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
	res, fetchAppsErr := c.do(http.MethodGet, fmt.Sprintf(appsPathPattern, groupID), api.RequestOptions{UseAuth: true})
	if fetchAppsErr != nil {
		return nil, fetchAppsErr
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, errGroupNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status error: %v", res.StatusCode)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var apps []App
	if err := dec.Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}
