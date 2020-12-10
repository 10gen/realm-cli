package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	appConfigJSON = "config.json"
)

var (
	maxDirectoryContainSearchDepth = 8
)

// AppData is MongoDB Realm application data
type AppData struct {
	ID   string `json:"client_app_id"`
	Name string `json:"name"`
}

// resolveAppData resolves the MongoDB Realm application based on the current working directory
// Empty data is successfully returned if this is called outside of a project directory
func resolveAppData(wd string) (AppData, error) {
	appDir, appDirOK, appDirErr := resolveAppDirectory(wd)
	if appDirErr != nil {
		return AppData{}, appDirErr
	}
	if !appDirOK {
		return AppData{}, nil
	}

	path := filepath.Join(appDir, appConfigJSON)

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

// resolveAppDirectory searches upwards from the current working directory
// for the root directory of a MongoDB Realm application project
func resolveAppDirectory(wd string) (string, bool, error) {
	wd, wdErr := filepath.Abs(wd)
	if wdErr != nil {
		return "", false, wdErr
	}

	for i := 0; i < maxDirectoryContainSearchDepth; i++ {
		path := filepath.Join(wd, appConfigJSON)
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
