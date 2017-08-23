// Package config provides functionality for interacting with stitch user
// authentication and configuration. Configuration data is stored as YAML in
// "~/.config/stitch".
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/10gen/stitch-cli/ui"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// Errors related to user authentication and configuration.
var (
	ErrNotLoggedIn    = errors.New("stitch: you are not logged in")
	ErrInvalidAPIKey  = errors.New("stitch: invalid API key")
	ErrNoDefaultGroup = errors.New("stitch: no default group")
)

var userConfig Config

func init() {
	home, _ := homedir.Dir()
	p := filepath.Join(home, ".config", "stitch")
	userConfig.path = p
	if _, err := os.Stat(p); err != nil {
		return
	}
	raw, err := ioutil.ReadFile(p)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(raw, &userConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stitch: failed to parse config at %q: %s", p, err)
		// no exit for non-logged-in operations here
		return
	}
}

// Config stores the user's login credentials and some metadata.
type Config struct {
	Name         string `yaml:"name"`
	Email        string `yaml:"email"`
	ID           string `yaml:"id"`
	Username     string `yaml:"username"`
	APIKey       string `yaml:"api_key"`
	RefreshToken string `yaml:"refresh_token"`
	DefaultGroup string `yaml:"default_group"`

	path string
}

func (c *Config) loggedIn() bool {
	return ValidAPIKey(c.APIKey)
}

func (c *Config) changeAndWrite(other Config) error {
	if other.Name != "" {
		c.Name = other.Name
	}
	if other.Email != "" {
		c.Email = other.Email
	}
	if other.ID != "" {
		c.ID = other.ID
	}
	if other.Username != "" {
		c.Username = other.Username
	}
	if other.APIKey != "" {
		c.APIKey = other.APIKey
	}
	if other.RefreshToken != "" {
		c.RefreshToken = other.RefreshToken
	}
	raw, _ := yaml.Marshal(c)
	err := ioutil.WriteFile(c.path, raw, 0600)
	return err
}

func (c *Config) logout() error {
	c.APIKey = ""
	c.RefreshToken = ""
	raw, _ := yaml.Marshal(c)
	err := ioutil.WriteFile(c.path, raw, 0600)
	return err
}

// User gets the user's configured login information.
func User() Config {
	return userConfig
}

// LoggedIn checks whether the local config has a logged in user.
func LoggedIn() bool {
	return userConfig.loggedIn()
}

// LogIn logs in using the given api key. This will delete the previously
// used api key and update the user's config.
func LogIn(apiKey string) error {
	if !ValidAPIKey(apiKey) {
		return ErrInvalidAPIKey
	}
	if LoggedIn() {
		ui.Ask("you are already logged in, this action will deauthenticate the existing user.\ncontinue?")
	}
	// TODO: actually get a new refresh token and other user fields, error if bad credentials
	refreshToken := "1234qwer0987poiu"
	err := userConfig.changeAndWrite(Config{
		APIKey:       apiKey,
		RefreshToken: refreshToken,
	})
	return err
}

// LogOut deletes the user's configured login credentials.
func LogOut() error {
	if !LoggedIn() {
		return ErrNotLoggedIn
	}
	return userConfig.logout()
}

// Fetch pull user data from stitch's services and updates the local config
// accordingly.
// This may log the user out if their login credentials are not valid.
func Fetch() error {
	// TODO
	return nil
}

// DefaultGroup gets the user's configured default group for use with stitch.
func DefaultGroup() (group string, err error) {
	group = userConfig.DefaultGroup
	if group == "" {
		err = ErrNoDefaultGroup
	}
	return
}

// ValidAPIKey locally checks if the given API key is valid.
func ValidAPIKey(apiKey string) bool {
	return len(apiKey) == 8 // TODO
}
