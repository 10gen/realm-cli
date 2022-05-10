package local

import (
	"bytes"
	"errors"
	"fmt"
	"os"
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
	Environment          realm.Environment                 `json:"environment,omitempty"`
	Environments         map[string]map[string]interface{} `json:"environments,omitempty"`
	Security             map[string]interface{}            `json:"security"`
	Hosting              map[string]interface{}            `json:"hosting,omitempty"`
	CustomUserDataConfig map[string]interface{}            `json:"custom_user_data_config"`
	Sync                 map[string]interface{}            `json:"sync"`
	Secrets              SecretsStructure                  `json:"secrets,omitempty"`
	AuthProviders        []map[string]interface{}          `json:"auth_providers,omitempty"`
	Functions            []map[string]interface{}          `json:"functions,omitempty"`
	Triggers             []map[string]interface{}          `json:"triggers,omitempty"`
	GraphQL              GraphQLStructure                  `json:"graphql,omitempty"`
	Services             []ServiceStructure                `json:"services,omitempty"`
	Values               []map[string]interface{}          `json:"values,omitempty"`
	LogForwarders        []map[string]interface{}          `json:"log_forwarders,omitempty"`
	DataAPIConfig        map[string]interface{}            `json:"data_api_config,omitempty"`
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

// Environment returns the local Realm app environment
func (a AppDataV1) Environment() realm.Environment {
	return a.AppStructureV1.Environment
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

	logForwarders, err := parseJSONFiles(filepath.Join(rootDir, NameLogForwarders))
	if err != nil {
		return err
	}
	a.LogForwarders = logForwarders

	dataAPIConfig, err := parseJSON(filepath.Join(rootDir, FileDataAPIConfig.String()))
	if err != nil {
		return err
	}
	if len(dataAPIConfig) > 0 {
		a.DataAPIConfig = dataAPIConfig
	}

	return nil
}

// ConfigData marshals the config data out to JSON
func (a AppDataV1) ConfigData() ([]byte, error) {
	temp := &struct {
		ConfigVersion        realm.AppConfigVersion `json:"config_version"`
		ID                   string                 `json:"app_id,omitempty"`
		Name                 string                 `json:"name"`
		Location             realm.Location         `json:"location"`
		DeploymentModel      realm.DeploymentModel  `json:"deployment_model"`
		Environment          realm.Environment      `json:"environment,omitempty"`
		Security             map[string]interface{} `json:"security"`
		Hosting              map[string]interface{} `json:"hosting,omitempty"`
		CustomUserDataConfig map[string]interface{} `json:"custom_user_data_config"`
		Sync                 map[string]interface{} `json:"sync"`
	}{
		ConfigVersion:        a.ConfigVersion(),
		ID:                   a.ID(),
		Name:                 a.Name(),
		Location:             a.Location(),
		DeploymentModel:      a.DeploymentModel(),
		Environment:          a.Environment(),
		Security:             a.Security,
		CustomUserDataConfig: a.CustomUserDataConfig,
		Sync:                 a.Sync,
	}
	return MarshalJSON(temp)
}

// WriteData will write the local Realm app data to disk
func (a *AppDataV1) WriteData(rootDir string) error {
	if err := writeSecrets(rootDir, a.Secrets); err != nil {
		return err
	}
	if err := writeEnvironments(rootDir, a.Environments); err != nil {
		return err
	}
	if err := writeValues(rootDir, a.Values); err != nil {
		return err
	}
	if err := writeGraphQL(rootDir, a.GraphQL); err != nil {
		return err
	}
	if err := writeServices(rootDir, a.Services); err != nil {
		return err
	}
	if err := writeFunctionsV1(rootDir, a.Functions); err != nil {
		return err
	}
	if err := writeAuthProviders(rootDir, a.AuthProviders); err != nil {
		return err
	}
	if err := writeTriggers(rootDir, a.Triggers); err != nil {
		return err
	}
	if err := writeLogForwarders(rootDir, a.LogForwarders); err != nil {
		return err
	}
	if err := writeDataAPIConfigV1(rootDir, a.DataAPIConfig); err != nil {
		return err
	}
	return nil
}

func writeFunctionsV1(rootDir string, functions []map[string]interface{}) error {
	dir := filepath.Join(rootDir, NameFunctions)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	for _, function := range functions {
		config, ok := function[NameConfig].(map[string]interface{})
		if !ok {
			return errors.New("error writing functions")
		}
		name, ok := config["name"].(string)
		if !ok {
			return errors.New("error writing functions")
		}
		data, err := MarshalJSON(config)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, name, FileConfig.String()),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
		src, ok := function[NameSource].(string)
		if !ok {
			return errors.New("error writing functions")
		}
		if err := WriteFile(
			filepath.Join(dir, name, FileSource.String()),
			0666,
			bytes.NewReader([]byte(src)),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeAuthProviders(rootDir string, authProviders []map[string]interface{}) error {
	dir := filepath.Join(rootDir, NameAuthProviders)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	for _, authProvider := range authProviders {
		data, err := MarshalJSON(authProvider)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, fmt.Sprintf("%s%s", authProvider["name"], extJSON)),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeDataAPIConfigV1(rootDir string, dataAPIConfig map[string]interface{}) error {
	if dataAPIConfig == nil {
		return nil
	}

	data, err := MarshalJSON(dataAPIConfig)
	if err != nil {
		return err
	}

	return WriteFile(
		filepath.Join(rootDir, FileDataAPIConfig.String()),
		0666,
		bytes.NewReader(data),
	)
}
