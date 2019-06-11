package models

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"fmt"
)

// DeploymentStatus is the enumeration of values which can be provided in a Deployment's status field
type DeploymentStatus string

// AppConfigFileName is the name of top-level config file describing the app
const AppConfigFileName string = "stitch.json"

// Default deployment settings
const (
	DefaultLocation        string = "US-VA"
	DefaultDeploymentModel string = "GLOBAL"

	// DeploymentStatusCreated indicates the app deployment has been created but is not in the job queue yet
	DeploymentStatusCreated DeploymentStatus = "created"

	// DeploymentStatusSuccessful indicates the app was successfully deployed
	DeploymentStatusSuccessful DeploymentStatus = "successful"

	// DeploymentStatusFailed indicates the app deployment failed
	DeploymentStatusFailed DeploymentStatus = "failed"

	// DeploymentStatusPending indicates the app deployment is in the job queue but has not yet started
	DeploymentStatusPending DeploymentStatus = "pending"
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

// AppDraft represents a Stitch App Draft
type AppDraft struct {
	ID string `json:"_id"`
}

// Deployment represents a Stitch Deployment
type Deployment struct {
	ID     string           `json:"_id"`
	Status DeploymentStatus `json:"status"`
}

// DraftDiff represents the diff of an AppDraft
type DraftDiff struct {
	Diffs            []string    `json:"diffs"`
	HostingFilesDiff HostingDiff `json:"hosting_files_diff"`
}

// HasChanges returns whether the DraftDiff contains any changes or not
func (d *DraftDiff) HasChanges() bool {
	return len(d.Diffs) != 0 ||
		len(d.HostingFilesDiff.Added) != 0 ||
		len(d.HostingFilesDiff.Deleted) != 0 ||
		len(d.HostingFilesDiff.Modified) != 0
}

// HostingDiff represents the hosting files section of a DraftDiff
type HostingDiff struct {
	Added    []string `json:"added"`
	Deleted  []string `json:"deleted"`
	Modified []string `json:"modified"`
}
