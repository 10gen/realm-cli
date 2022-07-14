package user

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/10gen/realm-cli/internal/telemetry"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	// DefaultProfile is the default profile name
	DefaultProfile = "default"

	// HostingAssetCacheDir is the hosting asset cache dir
	HostingAssetCacheDir = ".asset-cache"

	// ProfileType is the file type for profiles
	ProfileType = "yaml"

	envPrefix = "realm"

	extJSON = ".json"
)

// set of supported CLI user profile flags
const (
	FlagProfile = "profile"
	// TODO(REALMC-9249): add "[Learn more: http://docs.link]"
	FlagProfileUsage = `Specify your profile (Default value: "default")`

	FlagAtlasBaseURL      = "atlas-url"
	FlagAtlasBaseURLUsage = "specify the base Atlas server URL"

	FlagRealmBaseURL      = "realm-url"
	FlagRealmBaseURLUsage = "specify the base Realm server URL"

	defaultAtlasBaseURL = "https://cloud.mongodb.com"
	defaultRealmBaseURL = "https://realm.mongodb.com"
)

// Profile is the CLI profile
type Profile struct {
	Flags
	Name             string
	WorkingDirectory string

	dir string
	fs  afero.Fs
}

// Flags are the CLI profile flags
type Flags struct {
	AtlasBaseURL  string
	RealmBaseURL  string
	TelemetryMode telemetry.Mode
}

// NewDefaultProfile creates a new default CLI profile
func NewDefaultProfile() (*Profile, error) {
	return NewProfile(DefaultProfile)
}

// NewProfile creates a new CLI profile
func NewProfile(name string) (*Profile, error) {
	dir, dirErr := HomeDir()
	if dirErr != nil {
		return nil, fmt.Errorf("failed to create CLI profile: %w", dirErr)
	}

	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", wdErr)
	}

	return &Profile{
		Name:             name,
		dir:              dir,
		fs:               afero.NewOsFs(),
		WorkingDirectory: wd,
	}, nil
}

// Clear clears the specified CLI profile property
func (p Profile) Clear(name string) {
	p.SetString(name, "")
}

// SetString sets the specified CLI profile property
func (p Profile) SetString(name, value string) {
	viper.Set(p.propertyKey(name), value)
}

// GetString gets the specified CLI profile property
func (p Profile) GetString(name string) string {
	return viper.GetString(p.propertyKey(name))
}

func (p Profile) propertyKey(name string) string {
	return fmt.Sprintf("%s.%s", p.Name, name)
}

// Load loads the CLI profile
func (p Profile) Load() error {
	viper.SetConfigName(p.Name)
	viper.AddConfigPath(p.dir)
	viper.SetConfigPermissions(0600)
	viper.SetConfigType(ProfileType)

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil // proceed if profile doesn't exist
		}
		return fmt.Errorf("failed to load CLI profile: %s", err)
	}
	return nil
}

// Save saves the CLI profile
func (p *Profile) Save() error {
	exists, existsErr := afero.DirExists(p.fs, p.dir)
	if existsErr != nil {
		return fmt.Errorf("failed to save CLI profile: %s", existsErr)
	}

	if !exists {
		if err := p.fs.MkdirAll(p.dir, 0700); err != nil {
			return fmt.Errorf("failed to save CLI profile: %s", err)
		}
	}

	if err := viper.WriteConfigAs(p.Path()); err != nil {
		return fmt.Errorf("failed to save CLI profile: %s", err)
	}
	return nil
}

// ResolveFlags resolves the user profile flags
func (p *Profile) ResolveFlags() error {
	if p.Flags.TelemetryMode == telemetry.ModeEmpty {
		p.Flags.TelemetryMode = p.TelemetryMode()
	}
	p.SetString(keyTelemetryMode, string(p.Flags.TelemetryMode))

	if p.Flags.RealmBaseURL == "" {
		realmBaseURL := p.RealmBaseURL()
		if realmBaseURL == "" {
			realmBaseURL = defaultRealmBaseURL
		}
		p.Flags.RealmBaseURL = realmBaseURL
	}
	p.SetRealmBaseURL(p.Flags.RealmBaseURL)

	if p.Flags.AtlasBaseURL == "" {
		atlasBaseURL := p.AtlasBaseURL()
		if atlasBaseURL == "" {
			atlasBaseURL = defaultAtlasBaseURL
		}
		p.Flags.AtlasBaseURL = atlasBaseURL
	}
	p.SetAtlasBaseURL(p.Flags.AtlasBaseURL)

	return p.Save()
}

// Dir returns the CLI profile directory
func (p Profile) Dir() string {
	return p.dir
}

// Path returns the CLI profile filepath
func (p Profile) Path() string {
	return fmt.Sprintf("%s/%s.%s", p.dir, p.Name, ProfileType)
}

// set of supported CLI profile auth keys
const (
	keyPublicAPIKey  = "public_api_key"
	keyPrivateAPIKey = "private_api_key"
	keyUsername      = "username"
	keyPassword      = "password"
	keyAccessToken   = "access_token"
	keyRefreshToken  = "refresh_token"

	keyRealmBaseURL     = "realm_base_url"
	keyAtlasBaseURL     = "atlas_base_url"
	keyTelemetryMode    = "telemetry_mode"
	keyLastVersionCheck = "last_version_check"
)

// TelemetryMode gets the CLI profile telemetry mode
func (p Profile) TelemetryMode() telemetry.Mode {
	return telemetry.Mode(p.GetString(keyTelemetryMode))
}

// Credentials gets the CLI profile credentials
func (p Profile) Credentials() Credentials {
	return Credentials{
		PublicAPIKey:  p.GetString(keyPublicAPIKey),
		PrivateAPIKey: p.GetString(keyPrivateAPIKey),
		Username:      p.GetString(keyUsername),
		Password:      p.GetString(keyPassword),
	}
}

// SetCredentials sets the CLI profile credentials
func (p Profile) SetCredentials(creds Credentials) {
	p.SetString(keyPublicAPIKey, creds.PublicAPIKey)
	p.SetString(keyPrivateAPIKey, creds.PrivateAPIKey)
	p.SetString(keyUsername, creds.Username)
	p.SetString(keyPassword, creds.Password)
}

// ClearCredentials clears the CLI profile credentials
func (p Profile) ClearCredentials() {
	p.Clear(keyPublicAPIKey)
	p.Clear(keyPrivateAPIKey)
	p.Clear(keyUsername)
	p.Clear(keyPassword)
}

// Session gets the CLI profile session
func (p Profile) Session() Session {
	return Session{
		p.GetString(keyAccessToken),
		p.GetString(keyRefreshToken),
	}
}

// SetSession sets the CLI profile session
func (p Profile) SetSession(session Session) {
	p.SetString(keyAccessToken, session.AccessToken)
	p.SetString(keyRefreshToken, session.RefreshToken)
}

// ClearSession clears the CLI profile session
func (p Profile) ClearSession() {
	p.Clear(keyAccessToken)
	p.Clear(keyRefreshToken)
}

// RealmBaseURL gets the CLI profile Realm base url
func (p Profile) RealmBaseURL() string {
	return p.GetString(keyRealmBaseURL)
}

// SetRealmBaseURL sets the CLI profile Realm base url
func (p Profile) SetRealmBaseURL(realmBaseURL string) {
	p.SetString(keyRealmBaseURL, realmBaseURL)
}

// AtlasBaseURL gets the CLI profile Atlas base url
func (p Profile) AtlasBaseURL() string {
	return p.GetString(keyAtlasBaseURL)
}

// SetAtlasBaseURL sets the CLI profile Atlas base url
func (p Profile) SetAtlasBaseURL(realmBaseURL string) {
	p.SetString(keyAtlasBaseURL, realmBaseURL)
}

// LastVersionCheck gets the CLI profile last version check
func (p Profile) LastVersionCheck() time.Time {
	v := p.GetString(keyLastVersionCheck)

	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return time.Time{}
	}
	return t
}

// SetLastVersionCheck sets the CLI profile last version check
func (p Profile) SetLastVersionCheck(t time.Time) {
	p.SetString(keyLastVersionCheck, t.Format(time.RFC3339Nano))
}

// HostingAssetCachePath returns the CLI profile's hosting asset cache file path
func (p Profile) HostingAssetCachePath() string {
	return filepath.Join(p.dir, HostingAssetCacheDir, p.Name+extJSON)
}
