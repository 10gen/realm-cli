package local

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// AppData is the Realm app data
type AppData interface {
	ConfigData() ([]byte, error)
	ConfigVersion() realm.AppConfigVersion
	ID() string
	Name() string
	Location() realm.Location
	DeploymentModel() realm.DeploymentModel
	Environment() realm.Environment
	LoadData(rootDir string) error
	WriteData(rootDir string) error
}

// set of supported local names
const (
	extJS   = ".js"
	extJSON = ".json"

	// cli utilities
	NameDotMDB  = ".mdb"
	NameAppMeta = "app_meta"

	// app configs
	NameRealmConfig = "realm_config"
	NameConfig      = "config"
	NameStitch      = "stitch"

	// environments
	NameEnvironments = "environments"

	// auth
	NameAuth           = "auth"
	NameAuthProviders  = "auth_providers"
	NameCustomUserData = "custom_user_data"
	NameProviders      = "providers"

	// functions
	NameFunctions   = "functions"
	nameNodeModules = "node_modules"
	NameSource      = "source"
	NamePackageJSON = "package.json"

	// graphql
	NameGraphQL         = "graphql"
	NameCustomResolvers = "custom_resolvers"

	// hosting
	NameHosting  = "hosting"
	NameFiles    = "files"
	NameMetadata = "metadata"

	// services
	NameDataSources      = "data_sources"
	NameHTTPEndpoints    = "http_endpoints"
	NameIncomingWebhooks = "incoming_webhooks"
	NameDefaultRule      = "default_rule"
	NameRules            = "rules"
	NameSchema           = "schema"
	NameSchemas          = "schemas"
	NameServices         = "services"
	NameRelationships    = "relationships"

	// triggers
	NameTriggers = "triggers"

	// sync
	NameSync = "sync"

	// values
	NameSecrets = "secrets"
	NameValues  = "values"

	// log forwarders
	NameLogForwarders = "log_forwarders"

	// Data API Config
	NameDataAPIConfig = "data_api_config"
)

// set of supported local files
var (
	// cli utilities
	FileAppMeta = File{NameAppMeta, extJSON}

	// app configs
	FileRealmConfig = File{NameRealmConfig, extJSON}
	FileConfig      = File{NameConfig, extJSON}
	FileStitch      = File{NameStitch, extJSON}

	// auth
	FileCustomUserData = File{NameCustomUserData, extJSON}
	FileProviders      = File{NameProviders, extJSON}

	// data sources
	FileDefaultRule   = File{NameDefaultRule, extJSON}
	FileRules         = File{NameRules, extJSON}
	FileSchema        = File{NameSchema, extJSON}
	FileRelationships = File{NameRelationships, extJSON}

	// functions
	FileSource = File{NameSource, extJS}

	// values
	FileSecrets = File{NameSecrets, extJSON}

	// Data API Config
	FileDataAPIConfig = File{NameDataAPIConfig, extJSON}
)

// File is a local Realm app file
type File struct {
	Name string
	Ext  string
}

func (f File) String() string { return f.Name + f.Ext }

func walk(rootDir string, ignorePaths map[string]struct{}, fn func(file os.FileInfo, path string) error) error {
	if ignorePaths == nil {
		ignorePaths = map[string]struct{}{}
	}

	dw := directoryWalker{path: rootDir}
	if err := dw.walk(func(f os.FileInfo, p string) error {
		if _, ok := ignorePaths[f.Name()]; ok {
			return nil
		}
		if f.IsDir() {
			return walk(p, ignorePaths, fn)
		}
		return fn(f, p)
	}); err != nil {
		return err
	}
	return nil
}

type directoryWalker struct {
	path            string
	continueOnError bool
	failOnNotExist  bool
	onlyDirs        bool
	onlyFiles       bool
}

func (dw directoryWalker) walk(fn func(file os.FileInfo, path string) error) error {
	if _, err := os.Stat(dw.path); err != nil {
		if os.IsNotExist(err) && !dw.failOnNotExist {
			return nil
		}
		return err
	}
	files, filesErr := ioutil.ReadDir(dw.path)
	if filesErr != nil {
		return filesErr
	}
	for _, file := range files {
		if dw.onlyDirs && !file.IsDir() || dw.onlyFiles && file.IsDir() {
			continue
		}
		err := fn(file, filepath.Join(dw.path, file.Name()))
		if err != nil {
			if dw.continueOnError {
				continue
			}
			return err
		}
	}
	return nil
}

func readFile(path string) ([]byte, error) {
	return readFileWithOptions(path, false)
}

func readFileWithOptions(path string, failOnMissing bool) ([]byte, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) && !failOnMissing {
			return nil, nil
		}
		return nil, err
	}
	return ioutil.ReadFile(path)
}

func unmarshalJSON(data []byte, out interface{}) error {
	return unmarshalJSONWithOptions(data, out, false)
}

func unmarshalJSONWithOptions(data []byte, out interface{}, failOnEmpty bool) error {
	if len(data) == 0 {
		if failOnEmpty {
			return errors.New("no file contents")
		}
		return nil
	}
	return json.Unmarshal(data, out)
}

// AddAuthProvider adds an auth provider to the provided app data
func AddAuthProvider(appData AppData, name string, config map[string]interface{}) {
	switch ad := appData.(type) {
	case *AppStitchJSON:
		ad.AuthProviders = append(ad.AuthProviders, config)
	case *AppConfigJSON:
		ad.AuthProviders = append(ad.AuthProviders, config)
	case *AppRealmConfigJSON:
		if ad.Auth.Providers == nil {
			ad.Auth.Providers = map[string]interface{}{}
		}
		ad.Auth.Providers[name] = config
	}
}

// AddDataSource adds a data source to the app data
func AddDataSource(appData AppData, config map[string]interface{}) {
	switch ad := appData.(type) {
	case *AppStitchJSON:
		ad.Services = append(ad.Services, ServiceStructure{Config: config})
	case *AppConfigJSON:
		ad.Services = append(ad.Services, ServiceStructure{Config: config})
	case *AppRealmConfigJSON:
		ad.DataSources = append(ad.DataSources, DataSourceStructure{Config: config})
	}
}
