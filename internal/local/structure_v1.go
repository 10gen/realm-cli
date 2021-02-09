package local

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

// GraphQLStructure represents the Realm app graphql structure
type GraphQLStructure struct {
	Config          map[string]interface{}   `json:"config,omitempty"`
	CustomResolvers []map[string]interface{} `json:"custom_resolvers,omitempty"`
}

// SecretsStructure represents the Realm app secrets
type SecretsStructure struct {
	AuthProviders map[string]map[string]string `json:"auth_providers,omitempty"`
	Services      map[string]map[string]string `json:"services,omitempty"`
}

// ServiceStructure represents the Realm app service structure
type ServiceStructure struct {
	Config           map[string]interface{}   `json:"config,omitempty"`
	IncomingWebhooks []map[string]interface{} `json:"incoming_webhooks"`
	Rules            []map[string]interface{} `json:"rules"`
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
	if err := a.unmarshalSecrets(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalEnvironments(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalValues(rootDir); err != nil {
		return err
	}
	if err := a.unmarshalAuthProviders(rootDir); err != nil {
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

func (a *AppDataV1) unmarshalSecrets(rootDir string) error {
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

	var secrets SecretsStructure
	if err := json.Unmarshal(data, &secrets); err != nil {
		return err
	}
	a.Secrets = &secrets
	return nil
}

func (a *AppDataV1) unmarshalEnvironments(rootDir string) error {
	dir := filepath.Join(rootDir, NameEnvironments)

	environments := map[string]map[string]interface{}{}

	dw := directoryWalker{path: dir, onlyFiles: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		data, dataErr := readFile(path)
		if dataErr != nil {
			return dataErr
		}

		var out map[string]interface{}
		if err := unmarshalJSON(data, &out); err != nil {
			return err
		}

		environments[file.Name()] = out
		return nil
	}); walkErr != nil {
		return walkErr
	}

	if len(environments) > 0 {
		a.Environments = environments
	}
	return nil
}

func (a *AppDataV1) unmarshalAuthProviders(rootDir string) error {
	authProviders, err := unmarshalDirectoryFlat(filepath.Join(rootDir, NameAuthProviders))
	if err != nil {
		return err
	}
	a.AuthProviders = authProviders
	return nil
}

func (a *AppDataV1) unmarshalFunctions(rootDir string) error {
	functions, err := unmarshalFunctionsV1(filepath.Join(rootDir, NameFunctions))
	if err != nil {
		return err
	}
	a.Functions = functions
	return nil
}

func unmarshalFunctionsV1(path string) ([]map[string]interface{}, error) {
	var out []map[string]interface{}

	dw := directoryWalker{path: path, onlyDirs: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		if strings.Contains(path, nameNodeModules) {
			return nil // skip node_modules since we upload that as a single entity
		}

		cfg, cfgErr := readFile(filepath.Join(path, FileConfig.String()))
		if cfgErr != nil {
			return cfgErr
		}

		var config interface{}
		if err := unmarshalJSON(cfg, &config); err != nil {
			return err
		}

		src, srcErr := ioutil.ReadFile(filepath.Join(path, FileSource.String()))
		if srcErr != nil {
			return srcErr
		}

		o := map[string]interface{}{
			NameConfig: config,
			NameSource: string(src),
		}
		out = append(out, o)
		return nil
	}); walkErr != nil {
		return nil, walkErr
	}

	return out, nil
}

func (a *AppDataV1) unmarshalGraphQL(rootDir string) error {
	dir := filepath.Join(rootDir, NameGraphQL)
	if ok, err := fileExists(dir); err != nil {
		return err
	} else if !ok {
		return nil // graphql directory does not exist, continue
	}

	cfg, cfgErr := readFile(filepath.Join(dir, FileConfig.String()))
	if cfgErr != nil {
		return cfgErr
	}

	if err := unmarshalJSON(cfg, &a.GraphQL.Config); err != nil {
		return err
	}

	customResolvers := []map[string]interface{}{}
	dw := directoryWalker{path: filepath.Join(dir, NameCustomResolvers), onlyFiles: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		var out map[string]interface{}

		data, dataErr := readFile(path)
		if dataErr != nil {
			return dataErr
		}

		if err := unmarshalJSON(data, &out); err != nil {
			return err
		}

		customResolvers = append(customResolvers, out)
		return nil
	}); walkErr != nil {
		return walkErr
	}
	a.GraphQL.CustomResolvers = customResolvers
	return nil
}

func (a *AppDataV1) unmarshalServices(rootDir string) error {
	dir := filepath.Join(rootDir, NameServices)

	dw := directoryWalker{path: dir, onlyDirs: true}
	if walkErr := dw.walk(func(file os.FileInfo, path string) error {
		var out ServiceStructure

		cfg, cfgErr := readFile(filepath.Join(path, FileConfig.String()))
		if cfgErr != nil {
			return cfgErr
		}

		if err := unmarshalJSON(cfg, &out.Config); err != nil {
			return err
		}

		dirIncomingWebhooks := filepath.Join(path, NameIncomingWebhooks)
		if ok, err := fileExists(dirIncomingWebhooks); err != nil {
			return err
		} else if ok {
			incomingWebhooks, err := unmarshalFunctionsV1(dirIncomingWebhooks)
			if err != nil {
				return err
			}
			out.IncomingWebhooks = incomingWebhooks
		}

		dirRules := filepath.Join(path, NameRules)
		if ok, err := fileExists(dirRules); err != nil {
			return err
		} else if ok {
			rules, err := unmarshalDirectoryFlat(dirRules)
			if err != nil {
				return err
			}
			out.Rules = rules
		}

		a.Services = append(a.Services, out)
		return nil
	}); walkErr != nil {
		return walkErr
	}
	return nil
}

func (a *AppDataV1) unmarshalTriggers(rootDir string) error {
	triggers, err := unmarshalDirectoryFlat(filepath.Join(rootDir, NameTriggers))
	if err != nil {
		return err
	}
	a.Triggers = triggers
	return nil
}

func (a *AppDataV1) unmarshalValues(rootDir string) error {
	values, err := unmarshalDirectoryFlat(filepath.Join(rootDir, NameValues))
	if err != nil {
		return err
	}
	a.Values = values
	return nil
}
