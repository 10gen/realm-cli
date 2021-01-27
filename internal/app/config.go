package app

import (
	"errors"

	"github.com/10gen/realm-cli/internal/cloud/realm"
)

// set of known Realm app config fields
const (
	FieldDeploymentModel = "deployment_model"
	FieldName            = "name"
	FieldLocation        = "location"
)

// Config is the exported Realm app config
type Config struct {
	ConfigVersion        realm.AppConfigVersion `json:"config_version"`
	ID                   string                 `json:"app_id,omitempty"`
	Name                 string                 `json:"name"`
	Location             realm.Location         `json:"location"`
	DeploymentModel      realm.DeploymentModel  `json:"deployment_model"`
	Security             SecurityConfig         `json:"security"`
	CustomUserDataConfig CustomUserDataConfig   `json:"custom_user_data_config"`
	Sync                 SyncConfig             `json:"sync"`
}

// SecurityConfig ia the Realm app security config
type SecurityConfig struct {
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
}

// CustomUserDataConfig is the Realm app custom user data config
type CustomUserDataConfig struct {
	Enabled bool `json:"enabled"`
}

// SyncConfig is the Realm app sync config
type SyncConfig struct {
	DevelopmentModeEnabled bool `json:"development_mode_enabled"`
}

// ToDefaultConfig returns a default Realm app config based on the provided app
func ToDefaultConfig(app realm.App) Config {
	return Config{
		ConfigVersion:   realm.DefaultAppConfigVersion,
		ID:              app.ClientAppID,
		Name:            app.Name,
		Location:        app.Location,
		DeploymentModel: app.DeploymentModel,
	}
}

var (
	errUnknownConfigVersion = errors.New("unknown config version")
)

func configVersionFile(configVersion realm.AppConfigVersion) (File, error) {
	switch configVersion {
	case realm.AppConfigVersion20180301:
		return FileStitch, nil
	case realm.AppConfigVersion20200603:
		return FileConfig, nil
	case realm.AppConfigVersion20210101, realm.AppConfigVersionZero:
		return FileRealmConfig, nil //default
	}
	return File{}, errUnknownConfigVersion
}

func fileConfigVersion(config File) realm.AppConfigVersion {
	switch config.Name {
	case NameConfig:
		return realm.AppConfigVersion20200603
	case NameRealmConfig:
		return realm.AppConfigVersion20210101
	case NameStitch:
		return realm.AppConfigVersion20180301
	}
	return realm.AppConfigVersionZero
}
