package local

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

const (
	maxDirectoryContainSearchDepth = 8
)

const (
	// BackendPath is the relative path to write app contents to when we have templates
	BackendPath = "backend"
	// FrontendPath is the relative path to write frontend templates' contents to
	FrontendPath = "frontend"
)

func errFailedToParseAppConfig(path string) error {
	return errors.New("failed to parse app config at " + path)
}

func errFailedToFindApp(path string) error {
	return errors.New("failed to find app at " + path)
}

// App is the Realm app data represented on the local filesystem
type App struct {
	RootDir string
	Config  File
	AppMeta AppMeta
	AppData
}

// Option returns the Realm app data displayed as a selectable option
func (a App) Option() string {
	if a.AppData == nil {
		return a.RootDir
	}
	if a.AppData.ID() != "" {
		return a.AppData.ID()
	}
	return a.AppData.Name()
}

// NewApp returns a new local app
func NewApp(rootDir, clientAppID, name string, location realm.Location, deploymentModel realm.DeploymentModel, environment realm.Environment, configVersion realm.AppConfigVersion) App {
	return AsApp(rootDir, realm.App{
		ClientAppID: clientAppID,
		Name:        name,
		AppMeta: realm.AppMeta{
			Location:        location,
			DeploymentModel: deploymentModel,
			Environment:     environment,
		},
	}, configVersion)
}

// AsApp converts the realm.App into a local app
func AsApp(rootDir string, app realm.App, configVersion realm.AppConfigVersion) App {
	var appData AppData
	var config File
	switch configVersion {
	case realm.AppConfigVersion20180301:
		appData = &AppStitchJSON{AppDataV1{AppStructureV1{
			ConfigVersion:        configVersion,
			ID:                   app.ClientAppID,
			Name:                 app.Name,
			Location:             app.Location,
			DeploymentModel:      app.DeploymentModel,
			Environment:          app.Environment,
			CustomUserDataConfig: map[string]interface{}{"enabled": false},
			Sync:                 map[string]interface{}{"development_mode_enabled": false},
			Environments: map[string]map[string]interface{}{
				"development.json": {
					"values": map[string]interface{}{},
				},
				"no-environment.json": {
					"values": map[string]interface{}{},
				},
				"production.json": {
					"values": map[string]interface{}{},
				},
				"qa.json": {
					"values": map[string]interface{}{},
				},
				"testing.json": {
					"values": map[string]interface{}{},
				},
			},
			GraphQL: GraphQLStructure{
				Config: map[string]interface{}{
					"use_natural_pluralization": true,
				},
			},
		}}}
		config = FileStitch
	case realm.AppConfigVersion20200603:
		appData = &AppConfigJSON{AppDataV1{AppStructureV1{
			ConfigVersion:        configVersion,
			ID:                   app.ClientAppID,
			Name:                 app.Name,
			Location:             app.Location,
			DeploymentModel:      app.DeploymentModel,
			Environment:          app.Environment,
			CustomUserDataConfig: map[string]interface{}{"enabled": false},
			Sync:                 map[string]interface{}{"development_mode_enabled": false},
			Environments: map[string]map[string]interface{}{
				"development.json": {
					"values": map[string]interface{}{},
				},
				"no-environment.json": {
					"values": map[string]interface{}{},
				},
				"production.json": {
					"values": map[string]interface{}{},
				},
				"qa.json": {
					"values": map[string]interface{}{},
				},
				"testing.json": {
					"values": map[string]interface{}{},
				},
			},
			GraphQL: GraphQLStructure{
				Config: map[string]interface{}{
					"use_natural_pluralization": true,
				},
			},
		}}}
		config = FileConfig
	default:
		appData = &AppRealmConfigJSON{AppDataV2{AppStructureV2{
			ConfigVersion:   configVersion,
			ID:              app.ClientAppID,
			Name:            app.Name,
			Location:        app.Location,
			DeploymentModel: app.DeploymentModel,
			Environment:     app.Environment,
			Environments: map[string]map[string]interface{}{
				"development.json": {
					"values": map[string]interface{}{},
				},
				"no-environment.json": {
					"values": map[string]interface{}{},
				},
				"production.json": {
					"values": map[string]interface{}{},
				},
				"qa.json": {
					"values": map[string]interface{}{},
				},
				"testing.json": {
					"values": map[string]interface{}{},
				},
			},
			Auth: AuthStructure{
				CustomUserData: map[string]interface{}{"enabled": false},
				Providers:      map[string]interface{}{},
			},
			Sync: SyncStructure{Config: map[string]interface{}{"development_mode_enabled": false}},
			Functions: FunctionsStructure{
				Configs: []map[string]interface{}{},
				Sources: map[string]string{},
			},
			GraphQL: GraphQLStructure{
				Config: map[string]interface{}{
					"use_natural_pluralization": true,
				},
				CustomResolvers: []map[string]interface{}{},
			},
		}}}
		config = FileRealmConfig
	}
	return App{
		RootDir: rootDir,
		Config:  config,
		AppData: appData,
		AppMeta: AppMeta{
			ConfigVersion: configVersion,
			AppID:         app.ID,
			GroupID:       app.GroupID,
		},
	}
}

// Write writes the app data to disk
func (a App) Write() error {
	if a.AppData == nil {
		return nil
	}
	if err := a.WriteData(a.RootDir); err != nil {
		return err
	}
	return a.WriteConfig()
}

// WriteConfig writes the app config file to disk
func (a App) WriteConfig() error {
	if a.AppData == nil {
		return nil
	}
	data, err := a.AppData.ConfigData()
	if err != nil {
		return err
	}
	return WriteFile(filepath.Join(a.RootDir, a.Config.String()), 0666, bytes.NewReader(data))
}

// LoadApp will load the local app data and app config
func LoadApp(path string) (App, error) {
	app, appOK, appErr := FindApp(path)
	if appErr != nil {
		return App{}, appErr
	}
	if !appOK {
		return App{}, errFailedToFindApp(path)
	}

	if err := app.LoadData(app.RootDir); err != nil {
		return App{}, err
	}

	return app, nil
}

// LoadConfig will load the local app's config
func (a *App) LoadConfig() error {
	switch a.Config {
	case FileRealmConfig:
		a.AppData = &AppRealmConfigJSON{}
	case FileConfig:
		a.AppData = &AppConfigJSON{}
	case FileStitch:
		a.AppData = &AppStitchJSON{}
	default:
		return fmt.Errorf("invalid config file: %s", a.Config.String())
	}

	path := filepath.Join(a.RootDir, a.Config.String())

	data, dataErr := ioutil.ReadFile(path)
	if dataErr != nil {
		return errFailedToParseAppConfig(path)
	}

	if err := json.Unmarshal(data, a.AppData); err != nil {
		return errFailedToParseAppConfig(path)
	}
	return nil
}

// LoadAppMeta will load the local app's data from .mdb/app_meta.json
func (a *App) LoadAppMeta() error {
	path := filepath.Join(a.RootDir, NameDotMDB, FileAppMeta.String())
	data, dataErr := ioutil.ReadFile(path)
	if dataErr != nil {
		return errFailedToParseAppConfig(path)
	}

	if err := json.Unmarshal(data, &a.AppMeta); err != nil {
		return errFailedToParseAppConfig(path)
	}

	switch a.AppMeta.ConfigVersion {
	case realm.AppConfigVersion20180301:
		a.Config = FileStitch
	case realm.AppConfigVersion20200603:
		a.Config = FileConfig
	case realm.AppConfigVersion20210101:
		a.Config = FileRealmConfig
	default:
		return fmt.Errorf("invalid config version: %s", a.AppMeta.ConfigVersion.String())
	}

	return nil
}

var (
	allConfigFiles = []File{FileRealmConfig, FileConfig, FileStitch}
)

// FindApp searches upwards for the root of a Realm app project and
// returns the local app structure, a boolean indicating if the current
// working directory is part of a Realm app project, and any error that occurs
func FindApp(path string) (App, bool, error) {
	wd, wdErr := filepath.Abs(path)
	if wdErr != nil {
		return App{}, false, wdErr
	}

	for i := 0; i < maxDirectoryContainSearchDepth; i++ {
		if _, err := os.Stat(filepath.Join(wd, NameDotMDB, FileAppMeta.String())); err != nil {
			if os.IsNotExist(err) {
				// .mdb/app_meta.json may not exist, so we check whether a regular config file exists.
				if app, proceed, err := checkForConfigFile(wd); proceed || err != nil {
					return app, proceed, err
				}
			} else {
				return App{}, false, err
			}
		} else {
			// wd is the app root, so return the app no matter what
			app := App{RootDir: wd}
			if err := app.LoadAppMeta(); err != nil {
				return App{}, false, nil
			}
			_, err := os.Stat(filepath.Join(path, app.Config.String()))
			if err != nil {
				return App{}, false, err
			}
			if err := app.LoadConfig(); err != nil {
				return App{}, false, err
			}

			return app, true, nil
		}

		// No config file or cli file was found, so go up a directory
		if wd == "/" {
			break
		}
		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return App{}, false, nil
}

// Check for a config file at the path and determine whether or not to exit execution.
func checkForConfigFile(path string) (App, bool, error) {
	for _, config := range allConfigFiles {
		_, err := os.Stat(filepath.Join(path, config.String()))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return App{}, false, err
		}
		app := App{RootDir: path, Config: config}
		if err := app.LoadConfig(); err != nil {
			return App{}, false, err
		}

		if app.ConfigVersion() == realm.AppConfigVersionZero {
			continue
		}

		return app, true, nil
	}

	return App{}, false, nil // No valid config file
}
