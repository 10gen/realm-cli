package models

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"fmt"
)

// AppConfigFileName is the name of top-level config file describing the app
const AppConfigFileName string = "stitch.json"

// Default deployment settings
const (
	DefaultLocation        string = "US-VA"
	DefaultDeploymentModel string = "GLOBAL"
)

// App config field identifiers
const (
	AppIDField              string = "app_id"
	AppNameField            string = "name"
	AppLocationField        string = "location"
	AppDeploymentModelField string = "deployment_model"
)

const (
	prettyPrintPrefix = ""
	prettyPrintIndent = "    "
)

// AppInstanceData defines data pertaining to a specific deployment of a Stitch application
type AppInstanceData map[string]interface{}

// UnmarshalFile unmarshals data from a local config file into an AppInstanceData
func (aic *AppInstanceData) UnmarshalFile(path string) error {
	return readAndUnmarshalInto(json.Unmarshal, filepath.Join(path, AppConfigFileName), &aic)
}

// readAndUnmarshalInto unmarshals data from the given path into an interface{} using the provided marshalFn
func readAndUnmarshalInto(marshalFn func(in []byte, out interface{}) error, path string, out interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := marshalFn(data, out); err != nil {
		return fmt.Errorf("failed to parse %s: %s", path, err)
	}

	return nil
}

// MarshalFile writes the AppInstanceData to the AppConfigFileName at the provided path
func (aic *AppInstanceData) MarshalFile(path string) error {
	contents, err := json.MarshalIndent(aic, prettyPrintPrefix, prettyPrintIndent)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(path, AppConfigFileName), contents, 0600)
}

// AppID returns the app's Client App ID
func (aic AppInstanceData) AppID() string {
	appID, ok := aic[AppIDField].(string)
	if !ok {
		return ""
	}

	return appID
}

// AppName returns the app's user-defined name
func (aic AppInstanceData) AppName() string {
	appName, ok := aic[AppNameField].(string)
	if !ok {
		return ""
	}

	return appName
}

// AppLocation returns the app's target location
func (aic AppInstanceData) AppLocation() string {
	appLocation, ok := aic[AppLocationField].(string)
	if !ok {
		return DefaultLocation
	}

	return appLocation
}

// AppDeploymentModel returns the app's deployment model
func (aic AppInstanceData) AppDeploymentModel() string {
	appDeploymentModel, ok := aic[AppDeploymentModelField].(string)
	if !ok {
		return DefaultDeploymentModel
	}

	return appDeploymentModel
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
