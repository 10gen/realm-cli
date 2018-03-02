package models

import (
	"encoding/json"
	"path/filepath"

	"github.com/10gen/stitch-cli/utils"
)

// AppConfigFileName is the name of top-level config file describing the app
const AppConfigFileName string = "stitch.json"

// AppInstanceData defines data pertaining to a specific deployment of a Stitch application
type AppInstanceData struct {
	AppID   string `json:"app_id"`
	AppName string `json:"name"`
}

// UnmarshalFile unmarshals data from a local config file into an AppInstanceData
func (aic *AppInstanceData) UnmarshalFile(path string) error {
	return utils.ReadAndUnmarshalInto(json.Unmarshal, filepath.Join(path, AppConfigFileName), &aic)
}

// UserProfile holds basic metadata for a given user
type UserProfile struct {
	Roles []role `json:"roles"`
}

// AllGroupIDs returns all available group ids for a given user
func (pd *UserProfile) AllGroupIDs() []string {
	groupIDs := []string{}

	for _, role := range pd.Roles {
		if role.GroupID != "" {
			groupIDs = append(groupIDs, role.GroupID)
		}
	}

	return groupIDs
}

type role struct {
	GroupID string `json:"group_id"`
}

// App represents basic Stitch App data
type App struct {
	ID          string `json:"_id"`
	GroupID     string `json:"group_id"`
	ClientAppID string `json:"client_app_id"`
	Name        string `json:"name"`
}
