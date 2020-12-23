package profile

import (
	"fmt"
	"strings"

	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/telemetry"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

const (
	// DefaultProfile is the default profile name
	DefaultProfile = "default"

	envPrefix   = "realm"
	profileType = "yaml"

	// FlagProfile is the name of the profile flag
	FlagProfile = "profile"

	// FlagProfileShort is the short name of the profile flag
	FlagProfileShort = "i"

	// FlagProfileUsage is the usage text for the profile flag
	FlagProfileUsage = "this is the --profile, -p usage"

	// FlagAtlasBaseURL is the name of the atlas-url flag
	FlagAtlasBaseURL = "atlas-url"

	// FlagAtlasBaseURLUsage is the usage text for the atlas-url flag
	FlagAtlasBaseURLUsage = "specify the base Atlas server URL"

	// FlagRealmBaseURL is the name of the realm-url flag
	FlagRealmBaseURL = "realm-url"

	// FlagRealmBaseURLUsage is the usage text for the realm-url flag
	FlagRealmBaseURLUsage = "specify the base Realm server URL"

	defaultAtlasBaseURL = "https://cloud.mongodb.com"
	defaultRealmBaseURL = "https://realm.mongodb.com"
)

// Profile is the CLI profile
type Profile struct {
	Name          string
	AtlasBaseURL  string
	RealmBaseURL  string
	TelemetryMode telemetry.Mode

	dir string
	fs  afero.Fs
}

// NewDefaultProfile creates a new default CLI profile
func NewDefaultProfile() (*Profile, error) {
	return NewProfile(DefaultProfile)
}

// NewProfile creates a new CLI profile
func NewProfile(name string) (*Profile, error) {
	dir, dirErr := homeDir()
	if dirErr != nil {
		return nil, fmt.Errorf("failed to create CLI profile: %s", dirErr)
	}

	return &Profile{
		Name: name,
		dir:  dir,
		fs:   afero.NewOsFs(),
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
	viper.SetConfigType(profileType)

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

func (p Profile) ResolveFlags() error {
	if p.TelemetryMode == telemetry.ModeNil {
		p.TelemetryMode = telemetry.Mode(p.GetString(keyTelemetryMode))
	}
	p.SetString(keyTelemetryMode, string(p.TelemetryMode))

	if p.RealmBaseURL == "" {
		realmBaseURL := p.GetString(keyRealmBaseURL)
		if realmBaseURL == "" {
			realmBaseURL = defaultRealmBaseURL
		}
		p.RealmBaseURL = realmBaseURL
	}
	p.SetRealmBaseURL(p.RealmBaseURL)

	if p.AtlasBaseURL == "" {
		AtlasBaseURL := p.GetString(keyAtlasBaseURL)
		if AtlasBaseURL == "" {
			AtlasBaseURL = defaultAtlasBaseURL
		}
		p.AtlasBaseURL = AtlasBaseURL
	}
	p.SetAtlasBaseURL(p.AtlasBaseURL)

	return p.Save()
}

// Path returns the CLI profile filepath
func (p Profile) Path() string {
	return fmt.Sprintf("%s/%s.%s", p.dir, p.Name, profileType)
}

// set of supported CLI profile auth keys
const (
	keyPublicAPIKey  = "public_api_key"
	keyPrivateAPIKey = "private_api_key"
	keyAccessToken   = "access_token"
	keyRefreshToken  = "refresh_token"

	keyRealmBaseURL  = "realm_base_url"
	keyAtlasBaseURL  = "atlas_base_url"
	keyTelemetryMode = "telemetry_mode"
)

// User is the CLI profile user
type User struct {
	PublicAPIKey  string
	PrivateAPIKey string
}

// RedactedPrivateAPIKey returns the user's private API key with sensitive information redacted
func (user User) RedactedPrivateAPIKey() string {
	redact := func(s string) string {
		return strings.Repeat("*", len(s))
	}

	parts := strings.Split(user.PrivateAPIKey, "-")
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return redact(parts[0])
	default:
		lastIdx := len(parts) - 1

		out := make([]string, len(parts))
		for i := 0; i < lastIdx; i++ {
			out[i] = redact(parts[i])
		}
		out[lastIdx] = parts[lastIdx]

		return strings.Join(out, "-")
	}
}

// User gets the CLI profile user
func (p Profile) User() User {
	return User{
		p.GetString(keyPublicAPIKey),
		p.GetString(keyPrivateAPIKey),
	}
}

// SetUser sets the CLI profile user
func (p Profile) SetUser(publicAPIKey, privateAPIKey string) {
	p.SetString(keyPublicAPIKey, publicAPIKey)
	p.SetString(keyPrivateAPIKey, privateAPIKey)
}

// Session gets the CLI profile session
func (p Profile) Session() auth.Session {
	return auth.Session{
		p.GetString(keyAccessToken),
		p.GetString(keyRefreshToken),
	}
}

// SetSession sets the CLI profile session
func (p Profile) SetSession(accessToken, refreshToken string) {
	p.SetString(keyAccessToken, accessToken)
	p.SetString(keyRefreshToken, refreshToken)
}

// ClearSession clears the CLI profile session
func (p Profile) ClearSession() {
	p.Clear(keyAccessToken)
	p.Clear(keyRefreshToken)
}

// GetRealmBaseURL gets the CLI profile Realm base url
func (p Profile) GetRealmBaseURL() string {
	return p.GetString(keyRealmBaseURL)
}

// SetRealmBaseURL sets the CLI profile Realm base url
func (p Profile) SetRealmBaseURL(realmBaseURL string) {
	p.SetString(keyRealmBaseURL, realmBaseURL)
}

// GetAtlasBaseURL gets the CLI profile Atlas base url
func (p Profile) GetAtlasBaseURL() string {
	return p.GetString(keyAtlasBaseURL)
}

// SetAtlasBaseURL sets the CLI profile Atlas base url
func (p Profile) SetAtlasBaseURL(realmBaseURL string) {
	p.SetString(keyAtlasBaseURL, realmBaseURL)
}
