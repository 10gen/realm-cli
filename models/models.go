package models

import (
	"path/filepath"

	"github.com/10gen/stitch-cli/utils"

	"gopkg.in/yaml.v2"
)

const appInstanceDataFileName string = ".stitch"

// AppInstanceData defines data pertaining to a specific deployment of a Stitch application
type AppInstanceData struct {
	AppID string `yaml:"app_id"`
}

// UnmarshalFile unmarshals data from a local .stitch project file into an AppInstanceData
func (aic *AppInstanceData) UnmarshalFile(path string) error {
	return utils.ReadAndUnmarshalInto(yaml.Unmarshal, filepath.Join(path, appInstanceDataFileName), &aic)
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
}
