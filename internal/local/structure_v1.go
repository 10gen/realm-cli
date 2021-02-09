package local

import (
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// AppStructureV1 represents the v1 Realm app structure
type AppStructureV1 struct {
	ConfigVersion        realm.AppConfigVersion            `json:"config_version"`
	ID                   string                            `json:"app_id,omitempty"`
	Name                 string                            `json:"name"`
	Location             realm.Location                    `json:"location"`
	DeploymentModel      realm.DeploymentModel             `json:"deployment_model"`
	Environment          string                            `json:"environment,omitempty"`
	Environments         map[string]map[string]interface{} `json:"environments,omitempty"`
	Security             map[string]interface{}            `json:"security"`
	Hosting              map[string]interface{}            `json:"hosting,omitempty"`
	CustomUserDataConfig map[string]interface{}            `json:"custom_user_data_config"`
	Sync                 map[string]interface{}            `json:"sync"`
	Secrets              *SecretsStructure                 `json:"secrets,omitempty"`
	AuthProviders        []map[string]interface{}          `json:"auth_providers,omitempty"`
	Functions            []map[string]interface{}          `json:"functions,omitempty"`
	Triggers             []map[string]interface{}          `json:"triggers,omitempty"`
	GraphQL              GraphQLStructure                  `json:"graphql,omitempty"`
	Services             []ServiceStructure                `json:"services,omitempty"`
	Values               []map[string]interface{}          `json:"values,omitempty"`
}

// AppDataV1 is the v1 local Realm app data
type AppDataV1 struct {
	AppStructureV1
}

// ConfigVersion returns the local Realm app config version
func (a AppDataV1) ConfigVersion() realm.AppConfigVersion {
	return a.AppStructureV1.ConfigVersion
}

// ID returns the local Realm app id
func (a AppDataV1) ID() string {
	return a.AppStructureV1.ID
}

// Name returns the local Realm app name
func (a AppDataV1) Name() string {
	return a.AppStructureV1.Name
}

// Location returns the local Realm app location
func (a AppDataV1) Location() realm.Location {
	return a.AppStructureV1.Location
}

// DeploymentModel returns the local Realm app deployment model
func (a AppDataV1) DeploymentModel() realm.DeploymentModel {
	return a.AppStructureV1.DeploymentModel
}

// LoadData will load the local Realm app data
func (a *AppDataV1) LoadData(rootDir string) error {
	secrets, err := parseSecrets(rootDir)
	if err != nil {
		return err
	}
	a.Secrets = secrets

	environments, err := parseEnvironments(rootDir)
	if err != nil {
		return err
	}
	a.Environments = environments

	values, err := parseJSONFiles(filepath.Join(rootDir, NameValues))
	if err != nil {
		return err
	}
	a.Values = values

	authProviders, err := parseJSONFiles(filepath.Join(rootDir, NameAuthProviders))
	if err != nil {
		return err
	}
	a.AuthProviders = authProviders

	functions, err := parseFunctions(filepath.Join(rootDir, NameFunctions))
	if err != nil {
		return err
	}
	a.Functions = functions

	triggers, err := parseJSONFiles(filepath.Join(rootDir, NameTriggers))
	if err != nil {
		return err
	}
	a.Triggers = triggers

	graphql, ok, err := parseGraphQL(rootDir)
	if err != nil {
		return err
	} else if ok {
		a.GraphQL = graphql
	}

	services, err := parseServices(rootDir)
	if err != nil {
		return err
	}
	a.Services = services
	return nil
}
