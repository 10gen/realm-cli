package cli

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

var (
	maxDirectoryContainSearchDepth = 8
)

// AppData is the exported Realm app config data which it is identified by
type AppData struct {
	ID   string `json:"client_app_id,omitempty"`
	Name string `json:"name"`
}

// AppConfig is the exported Realm app config
type AppConfig struct {
	AppData
	ConfigVersion        realm.AppConfigVersion  `json:"config_version"`
	Location             realm.Location          `json:"location"`
	DeploymentModel      realm.DeploymentModel   `json:"deployment_model"`
	Security             AppSecurityConfig       `json:"security"`
	CustomUserDataConfig AppCustomUserDataConfig `json:"custom_user_data_config"`
	Sync                 AppSyncConfig           `json:"sync"`
}

// AppSecurityConfig ia the Realm app security config
type AppSecurityConfig struct{}

// AppCustomUserDataConfig is the Realm app custom user data config
type AppCustomUserDataConfig struct {
	Enabled bool `json:"enabled"`
}

// AppSyncConfig is the Realm app sync config
type AppSyncConfig struct {
	DevelopmentModeEnabled bool `json:"development_mode_enabled"`
}

// ResolveAppData resolves the MongoDB Realm application based on the current working directory
// Empty data is successfully returned if this is called outside of a project directory
func ResolveAppData(wd string) (AppData, error) {
	appDir, appDirOK, appDirErr := ResolveAppDirectory(wd)
	if appDirErr != nil {
		return AppData{}, appDirErr
	}
	if !appDirOK {
		return AppData{}, nil
	}

	path := filepath.Join(appDir, realm.FileAppConfig)

	data, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return AppData{}, readErr
	}

	if len(data) == 0 {
		return AppData{}, fmt.Errorf("failed to read app data at %s", path)
	}

	var appData AppData
	if err := json.Unmarshal(data, &appData); err != nil {
		return AppData{}, err
	}
	return appData, nil
}

// ResolveAppDirectory searches upwards from the current working directory
// for the root directory of a MongoDB Realm application project
func ResolveAppDirectory(wd string) (string, bool, error) {
	wd, wdErr := filepath.Abs(wd)
	if wdErr != nil {
		return "", false, wdErr
	}

	for i := 0; i < maxDirectoryContainSearchDepth; i++ {
		path := filepath.Join(wd, realm.FileAppConfig)
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
