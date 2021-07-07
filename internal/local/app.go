package local

import (
	"bytes"
	"encoding/json"
	"errors"
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

// App is the Realm app data represented on the local filesystem
type App struct {
	RootDir string
	Config  File
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
	}
	return App{
		RootDir: rootDir,
		Config:  FileRealmConfig,
		AppData: appData,
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

// Load will load the entire local app's data
func (a *App) Load() error {
	if a.AppData == nil {
		if err := a.LoadConfig(); err != nil {
			return err
		}
	}
	if err := a.AppData.LoadData(a.RootDir); err != nil {
		return err
	}
	return nil
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

// LoadApp will load the local app data
func LoadApp(path string) (App, error) {
	app, appErr := LoadAppConfig(path)
	if appErr != nil {
		return App{}, appErr
	}
	if app.AppData != nil {
		if err := app.Load(); err != nil {
			return App{}, err
		}
	}
	return app, nil
}

// LoadAppConfig will load the local app config
func LoadAppConfig(path string) (App, error) {
	app, appOK, appErr := FindApp(path)
	if appErr != nil {
		return App{}, appErr
	}
	if !appOK {
		return App{}, nil
	}

	if err := app.LoadConfig(); err != nil {
		return App{}, err
	}
	return app, nil
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
		for _, config := range allConfigFiles {
			_, err := os.Stat(filepath.Join(wd, config.String()))
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return App{}, false, err
			}
			return App{RootDir: wd, Config: config}, true, nil
		}
		if wd == "/" {
			break
		}
		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return App{}, false, nil
}
