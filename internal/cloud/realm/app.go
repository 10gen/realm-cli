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
	appsPathPattern = adminAPI + "/groups/%s/apps"
)

func (c *client) GetAppsForUser() ([]App, error) {
	profile, err := c.GetAuthProfile()
	if err != nil {
		return nil, err
	}
	groupIDs := profile.AllGroupIDs()
	var arr []App
	for _, groupID := range groupIDs {
		apps, err := c.GetApps(groupID)
		if err != nil {
			// Request will fail if any GetApps call fails
			return nil, err
		}
		arr = append(arr, apps...)
	}
	return arr, nil
}

// GetApps fetches all Realm Apps associated with the given groupID
func (c *client) GetApps(groupID string) ([]App, error) {
	res, fetchAppsErr := c.do(http.MethodGet, fmt.Sprintf(appsPathPattern, groupID), api.RequestOptions{UseAuth: true})
	if fetchAppsErr != nil {
		return nil, fetchAppsErr
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
