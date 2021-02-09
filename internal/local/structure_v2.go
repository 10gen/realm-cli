package local

import (
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// AppStructureV2 represents the v2 Realm app structure
type AppStructureV2 struct {
	ConfigVersion         realm.AppConfigVersion            `json:"config_version"`
	ID                    string                            `json:"app_id,omitempty"`
	Name                  string                            `json:"name,omitempty"`
	Location              realm.Location                    `json:"location,omitempty"`
	DeploymentModel       realm.DeploymentModel             `json:"deployment_model,omitempty"`
	Environment           string                            `json:"environment,omitempty"`
	Environments          map[string]map[string]interface{} `json:"environments,omitempty"`
	AllowedRequestOrigins []string                          `json:"allowed_request_origins,omitempty"`
	Values                []map[string]interface{}          `json:"values,omitempty"`
	Auth                  *AuthStructure                    `json:"auth,omitempty"`
	Functions             *FunctionsStructure               `json:"functions,omitempty"`
	Triggers              []map[string]interface{}          `json:"triggers,omitempty"`
	DataSources           []DataSourceStructure             `json:"data_sources,omitempty"`
	HTTPEndpoints         []HTTPEndpointStructure           `json:"http_endpoints,omitempty"`
	Services              []ServiceStructure                `json:"services,omitempty"`
	GraphQL               *GraphQLStructure                 `json:"graphql,omitempty"`
	Hosting               map[string]interface{}            `json:"hosting,omitempty"`
	Sync                  *SyncStructure                    `json:"sync,omitempty"`
	Secrets               *SecretsStructure                 `json:"secrets,omitempty"`
}

// AuthStructure represents the v2 Realm app auth structure
type AuthStructure struct {
	Config         map[string]interface{} `json:"config,omitempty"`
	CustomUserData map[string]interface{} `json:"custom_user_data,omitempty"`
	Providers      map[string]interface{} `json:"providers,omitempty"`
}

// DataSourceStructure represents the v2 Realm app data source structure
type DataSourceStructure struct {
	Config map[string]interface{} `json:"config,omitempty"`
	// TODO(REALMC-8016): include latest rules/schema for a data source
}

// FunctionsStructure represents the v2 Realm app functions structure
type FunctionsStructure struct {
	Config map[string]interface{} `json:"config,omitempty"`
	SrcMap map[string]string      `json:"src_map,omitempty"`
}

// HTTPEndpointStructure represents the v2 Realm app http endpoint structure
type HTTPEndpointStructure struct {
	Config           map[string]interface{}   `json:"config,omitempty"`
	IncomingWebhooks []map[string]interface{} `json:"incoming_webhooks,omitempty"`
}

// SyncStructure represents the v2 Realm app sync structure
type SyncStructure struct {
	Config map[string]interface{} `json:"config,omitempty"`
}

// AppDataV2 is the v2 local Realm app data
type AppDataV2 struct {
	AppStructureV2
}

// ConfigVersion returns the local Realm app config version
func (a AppDataV2) ConfigVersion() realm.AppConfigVersion {
	return a.AppStructureV2.ConfigVersion
}

// ID returns the local Realm app id
func (a AppDataV2) ID() string {
	return a.AppStructureV2.ID
}

// Name returns the local Realm app name
func (a AppDataV2) Name() string {
	return a.AppStructureV2.Name
}

// Location returns the local Realm app location
func (a AppDataV2) Location() realm.Location {
	return a.AppStructureV2.Location
}

// DeploymentModel returns the local Realm app deployment model
func (a AppDataV2) DeploymentModel() realm.DeploymentModel {
	return a.AppStructureV2.DeploymentModel
}

// LoadData will load the local Realm app data
func (a *AppDataV2) LoadData(rootDir string) error {
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

	auth, err := parseAuth(rootDir)
	if err != nil {
		return err
	}
	a.Auth = auth

	sync, err := parseSync(rootDir)
	if err != nil {
		return err
	}
	a.Sync = sync

	functions, err := parseFunctionsV2(rootDir)
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
		a.GraphQL = &graphql
	}

	services, err := parseServices(rootDir)
	if err != nil {
		return err
	}
	a.Services = services

	dataSources, err := parseDataSources(rootDir)
	if err != nil {
		return err
	}
	a.DataSources = dataSources

	httpEndpoints, err := parseHTTPEndpoints(rootDir)
	if err != nil {
		return err
	}
	a.HTTPEndpoints = httpEndpoints

	return nil
}

func parseAuth(rootDir string) (*AuthStructure, error) {
	dir := filepath.Join(rootDir, NameAuth)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	config, err := parseJSON(filepath.Join(dir, FileConfig.String()))
	if err != nil {
		return nil, err
	}

	customUserData, err := parseJSON(filepath.Join(dir, FileCustomUserData.String()))
	if err != nil {
		return nil, err
	}

	providers, err := parseJSON(filepath.Join(dir, FileProviders.String()))
	if err != nil {
		return nil, err
	}

	return &AuthStructure{config, customUserData, providers}, nil
}

func parseFunctionsV2(rootDir string) (*FunctionsStructure, error) {
	// TODO(REALMC-7989): actually unmarshal the functions directory exported by 20210101
	return nil, nil
}

func parseDataSources(rootDir string) ([]DataSourceStructure, error) {
	var out []DataSourceStructure

	dw := directoryWalker{
		path:     filepath.Join(rootDir, NameDataSources),
		onlyDirs: true,
	}
	if err := dw.walk(func(file os.FileInfo, path string) error {
		config, err := parseJSON(filepath.Join(path, FileConfig.String()))
		if err != nil {
			return err
		}

		// TODO(REALMC-8016): include latest rules/schema for a data source
		// rules, err := parseJSONFiles(filepath.Join(path, NameRules))
		// if err != nil {
		// 	return err
		// }

		out = append(out, DataSourceStructure{config})
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func parseHTTPEndpoints(rootDir string) ([]HTTPEndpointStructure, error) {
	var out []HTTPEndpointStructure

	dw := directoryWalker{
		path:     filepath.Join(rootDir, NameHTTPEndpoints),
		onlyDirs: true,
	}
	if err := dw.walk(func(file os.FileInfo, path string) error {
		config, err := parseJSON(filepath.Join(path, FileConfig.String()))
		if err != nil {
			return err
		}

		webhooks, err := parseFunctions(filepath.Join(path))
		if err != nil {
			return err
		}

		out = append(out, HTTPEndpointStructure{config, webhooks})
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func parseSync(rootDir string) (*SyncStructure, error) {
	dir := filepath.Join(rootDir, NameSync)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	config, err := parseJSON(filepath.Join(dir, FileConfig.String()))
	if err != nil {
		return nil, err
	}
	return &SyncStructure{config}, nil
}
