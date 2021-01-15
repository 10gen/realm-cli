package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// set of exported json options
const (
	ExportedJSONPrefix = ""
	ExportedJSONIndent = "    "
)

// Data is the exported Realm app config data which it is identified by
type Data struct {
	ID   string `json:"client_app_id,omitempty"`
	Name string `json:"name"`
}

// Config is the exported Realm app config
type Config struct {
	Data
	ConfigVersion        realm.AppConfigVersion `json:"config_version"`
	Location             realm.Location         `json:"location"`
	DeploymentModel      realm.DeploymentModel  `json:"deployment_model"`
	Security             SecurityConfig         `json:"security"`
	CustomUserDataConfig CustomUserDataConfig   `json:"custom_user_data_config"`
	Sync                 SyncConfig             `json:"sync"`
}

// SecurityConfig ia the Realm app security config
type SecurityConfig struct{}

// CustomUserDataConfig is the Realm app custom user data config
type CustomUserDataConfig struct {
	Enabled bool `json:"enabled"`
}

// SyncConfig is the Realm app sync config
type SyncConfig struct {
	DevelopmentModeEnabled bool `json:"development_mode_enabled"`
}

// set of app structure filepath parts
const (
	FileConfig = "config.json"

	DirAuthProviders = "auth_providers"

	DirGraphQL                = "graphql"
	DirGraphQLCustomResolvers = DirGraphQL + "/custom_resolvers"
	FileGraphQLConfig         = DirGraphQL + "/config.json"
)

// FileAuthProvider creates the auth provider config filepath
func FileAuthProvider(name string) string {
	return fmt.Sprintf("%s/%s.json", DirAuthProviders, name)
}

// ResolveData resolves the MongoDB Realm application based on the current working directory
// Empty data is successfully returned if this is called outside of a project directory
func ResolveData(wd string) (Data, error) {
	appDir, appDirOK, appDirErr := ResolveDirectory(wd)
	if appDirErr != nil {
		return Data{}, appDirErr
	}
	if !appDirOK {
		return Data{}, nil
	}

	path := filepath.Join(appDir, FileConfig)

	data, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return Data{}, readErr
	}

	if len(data) == 0 {
		return Data{}, fmt.Errorf("failed to read app data at %s", path)
	}

	var appData Data
	if err := json.Unmarshal(data, &appData); err != nil {
		return Data{}, err
	}
	return appData, nil
}

// app directory search configuration
const (
	maxDirectoryContainSearchDepth = 8
)

// ResolveDirectory searches upwards from the current working directory
// for the root directory of a MongoDB Realm application project
func ResolveDirectory(wd string) (string, bool, error) {
	wd, wdErr := filepath.Abs(wd)
	if wdErr != nil {
		return "", false, wdErr
	}

	for i := 0; i < maxDirectoryContainSearchDepth; i++ {
		path := filepath.Join(wd, FileConfig)
		if _, err := os.Stat(path); err == nil {
			return filepath.Dir(path), true, nil
		}

		if wd == "/" {
			break
		}
		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return "", false, nil
}
