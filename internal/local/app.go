package local

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

const (
	maxDirectoryContainSearchDepth = 8
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

func (a App) String() string {
	if a.AppData == nil {
		return a.RootDir
	}
	if a.AppData.ID() != "" {
		return a.AppData.ID()
	}
	return a.AppData.Name()
}

// NewApp returns a new local app
func NewApp(rootDir, name string, location realm.Location, deploymentModel realm.DeploymentModel) App {
	return AsApp(rootDir, realm.App{
		Name: name,
		AppMeta: realm.AppMeta{
			Location:        location,
			DeploymentModel: deploymentModel,
		},
	})
}

// AsApp converts the realm.App into a local app
func AsApp(rootDir string, app realm.App) App {
	return App{
		RootDir: rootDir,
		Config:  FileConfig, // TODO(REALMC-7653): update default config file here
		AppData: &AppRealmConfigJSON{AppDataV2{AppStructureV2{
			ConfigVersion:   realm.AppConfigVersion20200603, // TODO(REALMC-7653): update default config version here
			ID:              app.ClientAppID,
			Name:            app.Name,
			Location:        app.Location,
			DeploymentModel: app.DeploymentModel,
		}}},
	}
}

// Data returns the local app's data
func (a App) Data() AppData { return a.AppData }

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
			ok, err := fileExists(filepath.Join(wd, config.String()))
			if err != nil {
				return App{}, false, err
			}
			if ok {
				return App{RootDir: wd, Config: config}, true, nil
			}
		}
		if wd == "/" {
			break
		}
		wd = filepath.Clean(filepath.Join(wd, ".."))
	}

	return App{}, false, nil
}
