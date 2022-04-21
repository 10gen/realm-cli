package local

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
	nameREADME   = "README"
)

func errFailedToParseAppConfig(path string) error {
	return errors.New("failed to parse app config at " + path)
}

func errFailedToParseAppMeta(path string) error {
	return errors.New("failed to parse app meta at " + path)
}

func errFailedToFindApp(path string) error {
	return errors.New("failed to find app at " + path)
}

// App is the Realm app data represented on the local filesystem
type App struct {
	RootDir string
	Config  File
	Meta    AppMeta
	AppData
}

// AppMeta represents the app's metadata
type AppMeta struct {
	GroupID       string                 `json:"group_id"`
	AppID         string                 `json:"app_id"`
	ConfigVersion realm.AppConfigVersion `json:"config_version"`
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

	var appMeta AppMeta
	if app.GroupID != "" && app.ID != "" {
		appMeta = AppMeta{
			ConfigVersion: configVersion,
			AppID:         app.ID,
			GroupID:       app.GroupID,
		}
	}

	return App{
		RootDir: rootDir,
		Config:  config,
		AppData: appData,
		Meta:    appMeta,
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

func (a *App) loadConfig() error {
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

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errFailedToParseAppConfig(path)
	}

	if err := json.Unmarshal(data, a.AppData); err != nil {
		return errFailedToParseAppConfig(path)
	}
	return nil
}

func (a *App) loadMeta() error {
	path := filepath.Join(a.RootDir, NameDotMDB, FileAppMeta.String())

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errFailedToParseAppMeta(path)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errFailedToParseAppMeta(path)
	}

	if err := json.Unmarshal(data, &a.Meta); err != nil {
		return errFailedToParseAppMeta(path)
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
	wd, err := filepath.Abs(path)
	if err != nil {
		return App{}, false, err
	}

	for i := 0; i < maxDirectoryContainSearchDepth; i++ {
		if app, ok, err := loadAppRoot(wd); err != nil {
			return App{}, false, err
		} else if ok {
			return app, true, nil
		}

		if wd == "/" {
			break
		}
		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return App{}, false, nil
}

func loadAppRoot(path string) (App, bool, error) {
	app := App{RootDir: path}

	if err := app.loadMeta(); err != nil {
		return App{}, false, err
	}

	switch app.Meta.ConfigVersion {
	case realm.AppConfigVersion20210101:
		app.Config = FileRealmConfig
	case realm.AppConfigVersion20200603:
		app.Config = FileConfig
	case realm.AppConfigVersion20180301:
		app.Config = FileStitch
	default:
		for _, config := range allConfigFiles {
			if _, err := os.Stat(filepath.Join(path, config.String())); err == nil {
				app.Config = config
			} else if !os.IsNotExist(err) {
				return App{}, false, err
			}
		}
	}

	// for loop found no matching config file
	if app.Config.Name == "" {
		return App{}, false, nil
	}

	if err := app.loadConfig(); err != nil {
		return App{}, false, err
	}

	// matching config file is not at project root (e.g. a nested config.json)
	if app.ConfigVersion() == realm.AppConfigVersionZero {
		return App{}, false, nil
	}

	return app, true, nil
}

// FindReadme finds the Realm app README file for the specified template
func FindReadme(projectPath, relativePath, templateID string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(projectPath, FrontendPath, templateID, nameREADME+"*"))
	if err != nil || len(matches) != 1 {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(matches[0]))
	return filepath.Join(relativePath, FrontendPath, templateID, nameREADME+ext), nil
}
