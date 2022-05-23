package local

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
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
	Environment           realm.Environment                 `json:"environment,omitempty"`
	Environments          map[string]map[string]interface{} `json:"environments,omitempty"`
	AllowedRequestOrigins []string                          `json:"allowed_request_origins,omitempty"`
	Values                []map[string]interface{}          `json:"values,omitempty"`
	Auth                  AuthStructure                     `json:"auth,omitempty"`
	Functions             FunctionsStructure                `json:"functions,omitempty"`
	Triggers              []map[string]interface{}          `json:"triggers,omitempty"`
	DataSources           []DataSourceStructure             `json:"data_sources,omitempty"`
	HTTPServices          []HTTPServiceStructure            `json:"http_endpoints,omitempty"`
	Endpoints             EndpointStructure                 `json:"endpoints,omitempty"`
	Services              []ServiceStructure                `json:"services,omitempty"`
	GraphQL               GraphQLStructure                  `json:"graphql,omitempty"`
	Hosting               map[string]interface{}            `json:"hosting,omitempty"`
	Sync                  SyncStructure                     `json:"sync,omitempty"`
	Secrets               SecretsStructure                  `json:"secrets,omitempty"`
	LogForwarders         []map[string]interface{}          `json:"log_forwarders,omitempty"`
	DataAPIConfig         map[string]interface{}            `json:"data_api_config,omitempty"`
}

// AuthStructure represents the v2 Realm app auth structure
type AuthStructure struct {
	CustomUserData map[string]interface{} `json:"custom_user_data,omitempty"`
	Providers      map[string]interface{} `json:"providers,omitempty"`
}

// DataSourceStructure represents the v2 Realm app data source structure
type DataSourceStructure struct {
	Config      map[string]interface{}   `json:"config,omitempty"`
	DefaultRule map[string]interface{}   `json:"default_rule,omitempty"`
	Rules       []map[string]interface{} `json:"rules,omitempty"`
}

// FunctionsStructure represents the v2 Realm app functions structure
type FunctionsStructure struct {
	Configs []map[string]interface{} `json:"config,omitempty"`
	Sources map[string]string        `json:"sources,omitempty"`
}

// HTTPServiceStructure represents the v2 Realm app http endpoint structure
type HTTPServiceStructure struct {
	Config           map[string]interface{}   `json:"config,omitempty"`
	IncomingWebhooks []map[string]interface{} `json:"incoming_webhooks,omitempty"`
	Rules            []map[string]interface{} `json:"rules,omitempty"`
}

// EndpointStructure represents the v2 Realm app http endpoint structure
type EndpointStructure struct {
	Configs []map[string]interface{} `json:"config,omitempty"`
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

// Environment returns the local Realm app environment
func (a AppDataV2) Environment() realm.Environment {
	return a.AppStructureV2.Environment
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
		a.GraphQL = graphql
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

	httpServices, err := parseHTTPServices(rootDir)
	if err != nil {
		return err
	}
	a.HTTPServices = httpServices

	endpoints, err := parseEndpointsV2(rootDir)
	if err != nil {
		return err
	}
	a.Endpoints = endpoints

	logForwarders, err := parseJSONFiles(filepath.Join(rootDir, NameLogForwarders))
	if err != nil {
		return err
	}
	a.LogForwarders = logForwarders

	dataAPIConfig, err := parseJSON(filepath.Join(rootDir, NameHTTPEndpoints, FileDataAPIConfig.String()))
	if err != nil {
		return err
	}
	if len(dataAPIConfig) > 0 {
		a.DataAPIConfig = dataAPIConfig
	}

	return nil
}

func parseAuth(rootDir string) (AuthStructure, error) {
	dir := filepath.Join(rootDir, NameAuth)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return AuthStructure{}, nil
		}
		return AuthStructure{}, err
	}

	customUserData, err := parseJSON(filepath.Join(dir, FileCustomUserData.String()))
	if err != nil {
		return AuthStructure{}, err
	}

	providers, err := parseJSON(filepath.Join(dir, FileProviders.String()))
	if err != nil {
		return AuthStructure{}, err
	}

	return AuthStructure{customUserData, providers}, nil
}

func parseFunctionsV2(rootDir string) (FunctionsStructure, error) {
	dir := filepath.Join(rootDir, NameFunctions)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return FunctionsStructure{}, nil
		}
		return FunctionsStructure{}, err
	}

	configs, err := parseJSONArray(filepath.Join(dir, FileConfig.String()))
	if err != nil {
		return FunctionsStructure{}, err
	}

	sources := map[string]string{}
	if err := walk(dir, map[string]struct{}{nameNodeModules: {}}, func(file os.FileInfo, path string) error {
		if filepath.Ext(path) != extJS {
			return nil // looking for javascript files
		}

		pathRelative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		sources[pathRelative] = string(data)
		return nil
	}); err != nil {
		return FunctionsStructure{}, err
	}

	return FunctionsStructure{configs, sources}, nil
}

// TODO (REALMC-10879): support endpoints in older config versions
func parseEndpointsV2(rootDir string) (EndpointStructure, error) {
	dir := filepath.Join(rootDir, NameHTTPEndpoints)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return EndpointStructure{}, nil
		}
		return EndpointStructure{}, err
	}

	configs, err := parseJSONArray(filepath.Join(dir, FileConfig.String()))
	if err != nil {
		return EndpointStructure{}, err
	}

	return EndpointStructure{configs}, nil
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

		defaultRule, err := parseJSON(filepath.Join(path, FileDefaultRule.String()))
		if err != nil {
			return err
		}

		var rules []map[string]interface{}

		dbs := directoryWalker{path: path, onlyDirs: true}
		if err := dbs.walk(func(db os.FileInfo, dbPath string) error {

			colls := directoryWalker{path: dbPath, onlyDirs: true}
			if err := colls.walk(func(coll os.FileInfo, collPath string) error {
				// A valid data sources folder contains at least one of:
				// - a rules.json file
				// - a pair of files, schema.json and relationships.json
				// If neither of these conditions are true (e.g. there is only a schema.json
				// file, or there are no files in the directory at all), we should error

				// If we are not using app schemas, a valid data sources folder should contain all of
				// these files and we should error otherwise

				rule, err := parseJSON(filepath.Join(collPath, FileRules.String()))
				if err != nil {
					return err
				}

				schemaBody, err := parseJSON(filepath.Join(collPath, FileSchema.String()))
				if err != nil {
					return err
				}

				relationships, err := parseJSON(filepath.Join(collPath, FileRelationships.String()))
				if err != nil {
					return err
				}

				if rule == nil {
					if schemaBody == nil || relationships == nil {
						return fmt.Errorf("collection dir %s should contain a rules.json file and/or both schema.json and relationships.json files", collPath)
					}

					rule = map[string]interface{}{
						"database":   db.Name(),
						"collection": coll.Name(),
					}
				} else {
					if len(rule) == 0 {
						return errors.New("rules file cannot be empty")
					}
				}

				if schemaBody != nil {
					rule[NameSchema] = schemaBody
				}
				if relationships != nil {
					rule[NameRelationships] = relationships
				}
				rules = append(rules, rule)

				return nil
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		out = append(out, DataSourceStructure{config, defaultRule, rules})
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func parseHTTPServices(rootDir string) ([]HTTPServiceStructure, error) {
	var out []HTTPServiceStructure

	dw := directoryWalker{
		path:     filepath.Join(rootDir, NameHTTPEndpoints),
		onlyDirs: true,
	}
	if err := dw.walk(func(file os.FileInfo, path string) error {
		config, err := parseJSON(filepath.Join(path, FileConfig.String()))
		if err != nil {
			return err
		}

		webhooks, err := parseFunctions(filepath.Join(path, NameIncomingWebhooks))
		if err != nil {
			return err
		}
		if webhooks == nil {
			webhooks = []map[string]interface{}{}
		}

		rules, err := parseJSONFiles(filepath.Join(path, NameRules))
		if err != nil {
			return err
		}
		if rules == nil {
			rules = []map[string]interface{}{}
		}

		out = append(out, HTTPServiceStructure{config, webhooks, rules})
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func parseSync(rootDir string) (SyncStructure, error) {
	dir := filepath.Join(rootDir, NameSync)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return SyncStructure{}, nil
		}
		return SyncStructure{}, err
	}

	config, err := parseJSON(filepath.Join(dir, FileConfig.String()))
	if err != nil {
		return SyncStructure{}, err
	}
	return SyncStructure{config}, nil
}

// ConfigData marshals the config data out to JSON
func (a AppDataV2) ConfigData() ([]byte, error) {
	temp := &struct {
		ConfigVersion         realm.AppConfigVersion `json:"config_version"`
		ID                    string                 `json:"app_id,omitempty"`
		Name                  string                 `json:"name,omitempty"`
		Location              realm.Location         `json:"location,omitempty"`
		DeploymentModel       realm.DeploymentModel  `json:"deployment_model,omitempty"`
		Environment           realm.Environment      `json:"environment,omitempty"`
		AllowedRequestOrigins []string               `json:"allowed_request_origins,omitempty"`
	}{
		ConfigVersion:         a.ConfigVersion(),
		ID:                    a.ID(),
		Name:                  a.Name(),
		Location:              a.Location(),
		DeploymentModel:       a.DeploymentModel(),
		Environment:           a.Environment(),
		AllowedRequestOrigins: a.AllowedRequestOrigins,
	}
	return MarshalJSON(temp)
}

// WriteData will write the local Realm app data to disk
func (a AppDataV2) WriteData(rootDir string) error {
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
	if err := writeFunctionsV2(rootDir, a.Functions); err != nil {
		return err
	}
	if err := writeAuth(rootDir, a.Auth); err != nil {
		return err
	}
	if err := writeSync(rootDir, a.Sync); err != nil {
		return err
	}
	if err := writeDataSources(rootDir, a.DataSources); err != nil {
		return err
	}
	if err := writeHTTPServices(rootDir, a.HTTPServices); err != nil {
		return err
	}
	if err := writeEndpoints(rootDir, a.Endpoints); err != nil {
		return err
	}
	if err := writeTriggers(rootDir, a.Triggers); err != nil {
		return err
	}
	if err := writeLogForwarders(rootDir, a.LogForwarders); err != nil {
		return err
	}
	if err := writeDataAPIConfigV2(rootDir, a.DataAPIConfig); err != nil {
		return err
	}
	return nil
}

func writeFunctionsV2(rootDir string, functions FunctionsStructure) error {
	dir := filepath.Join(rootDir, NameFunctions)
	data, err := MarshalJSON(functions.Configs)
	if err != nil {
		return err
	}

	if err := WriteFile(
		filepath.Join(dir, FileConfig.String()),
		0666,
		bytes.NewReader(data),
	); err != nil {
		return err
	}

	for path, src := range functions.Sources {
		if err := WriteFile(
			filepath.Join(dir, path),
			0666,
			bytes.NewReader([]byte(src)),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeAuth(rootDir string, auth AuthStructure) error {
	dir := filepath.Join(rootDir, NameAuth)

	if auth.Providers != nil {
		data, err := MarshalJSON(auth.Providers)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, FileProviders.String()),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}

	if auth.CustomUserData != nil {
		data, err := MarshalJSON(auth.CustomUserData)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, FileCustomUserData.String()),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
	}

	return nil
}

func writeSync(rootDir string, sync SyncStructure) error {
	if sync.Config == nil {
		return nil
	}

	data, err := MarshalJSON(sync.Config)
	if err != nil {
		return err
	}

	return WriteFile(
		filepath.Join(rootDir, NameSync, FileConfig.String()),
		0666,
		bytes.NewReader(data),
	)
}

func writeDataSources(rootDir string, dataSources []DataSourceStructure) error {
	dir := filepath.Join(rootDir, NameDataSources)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	for _, ds := range dataSources {
		name, ok := ds.Config["name"].(string)
		if !ok {
			return errors.New("error writing datasources")
		}

		// Config
		config, err := MarshalJSON(ds.Config)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, name, FileConfig.String()),
			0666,
			bytes.NewReader(config),
		); err != nil {
			return err
		}

		// Default Rule
		if ds.DefaultRule != nil {
			defaultRule, err := MarshalJSON(ds.DefaultRule)
			if err != nil {
				return err
			}
			if err := WriteFile(
				filepath.Join(dir, name, FileDefaultRule.String()),
				0666,
				bytes.NewReader(defaultRule),
			); err != nil {
				return err
			}
		}

		// Rules
		for _, rule := range ds.Rules {
			ruleTemp := map[string]interface{}{}
			for k, v := range rule {
				ruleTemp[k] = v
			}
			delete(ruleTemp, NameSchema)
			delete(ruleTemp, NameRelationships)
			dataRule, err := MarshalJSON(ruleTemp)
			if err != nil {
				return err
			}
			var database, collection string
			if db, ok := rule["database"]; ok {
				if db, ok := db.(string); ok {
					database = db
				}
			}
			if coll, ok := rule["collection"]; ok {
				if coll, ok := coll.(string); ok {
					collection = coll
				}
			}
			ruleDir := filepath.Join(dir, name, database, collection)
			if err := WriteFile(
				filepath.Join(ruleDir, FileRules.String()),
				0666,
				bytes.NewReader(dataRule),
			); err != nil {
				return err
			}

			schema, ok := rule[NameSchema]
			if ok {
				data, err := MarshalJSON(schema)
				if err != nil {
					return err
				}
				if err := WriteFile(
					filepath.Join(ruleDir, FileSchema.String()),
					0666,
					bytes.NewReader(data),
				); err != nil {
					return err
				}
			}

			relationships, ok := rule[NameRelationships]
			if ok {
				data, err := MarshalJSON(relationships)
				if err != nil {
					return err
				}
				if err := WriteFile(
					filepath.Join(ruleDir, FileRelationships.String()),
					0666,
					bytes.NewReader(data),
				); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func writeHTTPServices(rootDir string, httpServices []HTTPServiceStructure) error {
	dir := filepath.Join(rootDir, NameHTTPEndpoints)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	for _, httpService := range httpServices {
		nameHTTPEndpoint, ok := httpService.Config["name"].(string)
		if !ok {
			return errors.New("error writing http endpoints")
		}
		data, err := MarshalJSON(httpService.Config)
		if err != nil {
			return err
		}
		if err := WriteFile(
			filepath.Join(dir, nameHTTPEndpoint, FileConfig.String()),
			0666,
			bytes.NewReader(data),
		); err != nil {
			return err
		}
		for _, webhook := range httpService.IncomingWebhooks {
			src, ok := webhook[NameSource].(string)
			if !ok {
				return errors.New("error writing http endpoints")
			}
			name, ok := webhook["name"].(string)
			if !ok {
				return errors.New("error writing http endpoints")
			}
			dirHTTPEndpoint := filepath.Join(dir, nameHTTPEndpoint, NameIncomingWebhooks, name)
			webhookTemp := map[string]interface{}{}
			for k, v := range webhook {
				webhookTemp[k] = v
			}
			delete(webhookTemp, NameSource)
			config, err := MarshalJSON(webhookTemp)
			if err != nil {
				return err
			}
			if err := WriteFile(
				filepath.Join(dirHTTPEndpoint, FileConfig.String()),
				0666,
				bytes.NewReader(config),
			); err != nil {
				return err
			}
			if err := WriteFile(
				filepath.Join(dirHTTPEndpoint, FileSource.String()),
				0666,
				bytes.NewReader([]byte(src)),
			); err != nil {
				return err
			}
		}
		for _, rule := range httpService.Rules {
			data, err := MarshalJSON(rule)
			if err != nil {
				return err
			}
			if err := WriteFile(
				filepath.Join(dir, nameHTTPEndpoint, NameRules, fmt.Sprintf("%s%s", rule["name"], extJSON)),
				0666,
				bytes.NewReader(data),
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeDataAPIConfigV2(rootDir string, dataAPIConfig map[string]interface{}) error {
	if dataAPIConfig == nil {
		return nil
	}

	data, err := MarshalJSON(dataAPIConfig)
	if err != nil {
		return err
	}

	return WriteFile(
		filepath.Join(rootDir, NameHTTPEndpoints, FileDataAPIConfig.String()),
		0666,
		bytes.NewReader(data),
	)
}
