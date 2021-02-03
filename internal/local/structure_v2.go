package local

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// AppStructureV2 represents the v2 Realm app structure
type AppStructureV2 struct {
	ConfigVersion         realm.AppConfigVersion   `json:"config_version"`
	ID                    string                   `json:"app_id,omitempty"`
	Name                  string                   `json:"name,omitempty"`
	Location              realm.Location           `json:"location,omitempty"`
	DeploymentModel       realm.DeploymentModel    `json:"deployment_model,omitempty"`
	Environment           string                   `json:"environment,omitempty"`
	AllowedRequestOrigins []string                 `json:"allowed_request_origins,omitempty"`
	Secrets               map[string]interface{}   `json:"secrets,omitempty"`
	Values                []map[string]interface{} `json:"values,omitempty"`
	Auth                  *AuthStructure           `json:"auth,omitempty"`
	Functions             *FunctionsStructure      `json:"functions,omitempty"`
	Triggers              []map[string]interface{} `json:"triggers,omitempty"`
	Services              []ServiceStructure       `json:"services,omitempty"`
	GraphQL               *GraphQLStructure        `json:"graphql,omitempty"`
	Sync                  *SyncStructure           `json:"sync,omitempty"`
}

// AuthStructure represents the v2 Realm app auth structure
type AuthStructure struct {
	Config    map[string]interface{}   `json:"config,omitempty"`
	Providers []map[string]interface{} `json:"providers,omitempty"`
}

// FunctionsStructure represents the v2 Realm app functions structure
type FunctionsStructure struct {
	Config map[string]interface{} `json:"config,omitempty"`
	SrcMap map[string]string      `json:"src_map,omitempty"`
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
	if err := a.unmarshalSecrets(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalValues(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalAuth(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalSync(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalFunctions(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalTriggers(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalGraphQL(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalServices(rootDir); err != nil {
		return err
	}
	return nil
}

func (a *AppDataV2) unmarshalSecrets(rootDir string) error {
	path := filepath.Join(rootDir, FileSecrets.String())
	if ok, err := fileExists(path); err != nil {
		return err
	} else if !ok {
		return nil // if secrets.json does not exist, continue
	}

	data, dataErr := readFile(path)
	if dataErr != nil {
		return dataErr
	}
	return json.Unmarshal(data, &a.Secrets)
}

func (a *AppDataV2) unmarshalAuth(rootDir string) error {
	dir := filepath.Join(rootDir, NameAuth)
	if ok, err := fileExists(dir); err != nil {
		return err
	} else if !ok {
		return nil // auth directory does not exist, continue
	}

	var auth AuthStructure

	cfg, cfgErr := readFile(filepath.Join(dir, FileConfig.String()))
	if cfgErr != nil {
		return cfgErr
	}

	if err := json.Unmarshal(cfg, &auth.Config); err != nil {
		return err
	}

	authProviders, err := unmarshalDirectoryFlat(filepath.Join(dir, NameAuthProviders))
	if err != nil {
		return err
	}
	auth.Providers = authProviders

	if len(auth.Config) > 0 || len(auth.Providers) > 0 {
		a.Auth = &auth
	}
	return nil
}

func (a *AppDataV2) unmarshalFunctions(rootDir string) error {
	// TODO(REALMC-7653): actually unmarshal the functions directory exported by 20210101
	return nil
}

func (a *AppDataV2) unmarshalGraphQL(rootDir string) error {
	dir := filepath.Join(rootDir, NameGraphQL)
	if ok, err := fileExists(dir); err != nil {
		return err
	} else if !ok {
		return nil // graphql directory does not exist, continue
	}

	var graphql GraphQLStructure

	cfg, cfgErr := readFile(filepath.Join(dir, FileConfig.String()))
	if cfgErr != nil {
		return cfgErr
	}

	if err := json.Unmarshal(cfg, &graphql.Config); err != nil {
		return err
	}

	dw := directoryWalker{path: filepath.Join(dir, NameCustomResolvers), onlyFiles: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		var out map[string]interface{}

		data, dataErr := readFile(filepath.Join(dir, FileConfig.String()))
		if dataErr != nil {
			return dataErr
		}

		if err := json.Unmarshal(data, &out); err != nil {
			return err
		}

		graphql.CustomResolvers = append(graphql.CustomResolvers, out)
		return nil
	}); walkErr != nil {
		return walkErr
	}

	if len(graphql.Config) > 0 || len(graphql.CustomResolvers) > 0 {
		a.GraphQL = &graphql
	}
	return nil
}

func (a *AppDataV2) unmarshalServices(rootDir string) error {
	dir := filepath.Join(rootDir, NameServices)

	dw := directoryWalker{path: dir, onlyDirs: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		var service ServiceStructure

		cfg, cfgErr := readFile(filepath.Join(dir, FileConfig.String()))
		if cfgErr != nil {
			return cfgErr
		}

		if err := json.Unmarshal(cfg, &service.Config); err != nil {
			return err
		}

		dirIncomingWebhooks := filepath.Join(dir, NameIncomingWebhooks)
		if ok, err := fileExists(dirIncomingWebhooks); err != nil {
			return err
		} else if ok {
			incomingWebhooks, err := unmarshalFunctionsV1(dirIncomingWebhooks)
			if err != nil {
				return err
			}
			service.IncomingWebhooks = incomingWebhooks
		}

		dirRules := filepath.Join(dir, NameRules)
		if ok, err := fileExists(dirRules); err != nil {
			return err
		} else if ok {
			rules, err := unmarshalDirectoryFlat(dirRules)
			if err != nil {
				return err
			}
			service.Rules = rules
		}

		a.Services = append(a.Services, service)
		return nil
	}); walkErr != nil {
		return walkErr
	}
	return nil
}

func (a *AppDataV2) unmarshalSync(rootDir string) error {
	dir := filepath.Join(rootDir, NameSync)
	if ok, err := fileExists(dir); err != nil {
		return err
	} else if !ok {
		return nil // sync directory does not exist, continue
	}

	var sync SyncStructure

	cfg, cfgErr := readFile(filepath.Join(dir, FileConfig.String()))
	if cfgErr != nil {
		return cfgErr
	}

	if err := json.Unmarshal(cfg, &sync.Config); err != nil {
		return err
	}

	if len(sync.Config) > 0 {
		a.Sync = &sync
	}
	return nil
}

func (a *AppDataV2) unmarshalTriggers(rootDir string) error {
	triggers, err := unmarshalDirectoryFlat(filepath.Join(rootDir, NameTriggers))
	if err != nil {
		return err
	}
	a.Triggers = triggers
	return nil
}

func (a *AppDataV2) unmarshalValues(rootDir string) error {
	values, err := unmarshalDirectoryFlat(filepath.Join(rootDir, NameValues))
	if err != nil {
		return err
	}
	a.Values = values
	return nil
}
