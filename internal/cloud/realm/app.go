package realm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/10gen/realm-cli/internal/utils/api"
)

// UserProfile contains all of the roles associated with a user
type UserProfile struct {
	Roles []Role `json:"roles"`
}

// Role contains a GrouID field that maps to a project ID
type Role struct {
	GroupID string `json:"group_id"`
}

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

// FilterGroupIDsFromUserProfile returns all available group ids for a given user
func (pd *UserProfile) FilterGroupIDsFromUserProfile(projectFlag string) []string {
	groupIDs := []string{}
	set := make(map[string]bool)
	for _, role := range pd.Roles {
		// Break early if we find the groupID that matches the project flag
		if len(projectFlag) == 0 && !set[role.GroupID] {
			groupIDs = append(groupIDs, role.GroupID)
			set[role.GroupID] = true
			continue
		}

		if role.GroupID == projectFlag {
			return []string{role.GroupID}
		}
	}

	return groupIDs
}

func (c *client) GetUserProfile() (UserProfile, error) {
	opts, _ := api.JSONRequestOptions(nil)

	res, resErr := c.do(http.MethodGet, appsByGroupIDEndpoint, opts)
	if resErr != nil {
		return UserProfile{}, resErr
	}

	if res.StatusCode != http.StatusOK {
		return UserProfile{}, UnmarshalServerError(res)
	}

	dec := json.NewDecoder(res.Body)
	defer res.Body.Close()

	var profile UserProfile
	if err := dec.Decode(&profile); err != nil {
		return UserProfile{}, err
	}

	return profile, nil
}

func (c *client) FindProjectAppByClientAppID(groupIDs []string, appFlag string) ([]App, error) {
	var appArr []App
	appArr = make([]App, 0)
	for _, groupID := range groupIDs {
		apps, err := c.FetchAppsByGroupID(groupID)
		if err != nil && err != errGroupNotFound {
			//TODO REALMC-7156: should we fail silently or hard here
			return nil, err
		}
		appSet := make(map[string]bool)
		for _, appElem := range apps {
			if len(appFlag) > 0 {
				if appElem.Name == appFlag {
					appArr = append(appArr, *appElem)
					// Greedily return the first app that matches if --app flag provided
					// Note: User could have multiple projects with an app of the same name]
					return appArr, nil
				}
				continue
			}
			if !appSet[appElem.Name] {
				appArr = append(appArr, *appElem)
				appSet[appElem.Name] = true
			}

		}
	}

	return appArr, nil
}

func (c *client) FetchAppsByGroupID(groupID string) ([]*App, error) {
	opts, _ := api.JSONRequestOptions(nil)

	res, err := c.do(http.MethodGet, fmt.Sprintf(appsByGroupIDEndpoint, groupID), opts)

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
	var apps []*App
	if err := dec.Decode(&apps); err != nil {
		return nil, err
	}
	return apps, nil
}
