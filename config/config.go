package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

var (
	ErrAlreadyLoggedIn = errors.New("stitch: you are already logged in.")
	ErrNotLoggedIn     = errors.New("stitch: you are not logged in.")
	ErrInvalidAPIKey   = errors.New("stitch: invalid API key.")
)

var userConfig config

var Yes bool // set by -y/--yes in commands/command.go

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

type config struct {
	ApiKey       string `yaml:"api_key"`
	RefreshToken string `yaml:"refresh_token"`

	path string
}

func (c *config) loggedIn() bool {
	return ValidApiKey(c.ApiKey)
}

func (c *config) changeAndWrite(other config) error {
	if other.ApiKey != "" {
		c.ApiKey = other.ApiKey
	}
	if other.RefreshToken != "" {
		c.RefreshToken = other.RefreshToken
	}
	raw, _ := yaml.Marshal(c)
	err := ioutil.WriteFile(c.path, raw, 0600)
	return err
}

func (c *config) logout() error {
	*c = config{path: c.path} // zero out fields
	raw, _ := yaml.Marshal(c)
	err := ioutil.WriteFile(c.path, raw, 0600)
	return err
}

// LoggedIn checks whether the local config has a logged in user.
func LoggedIn() bool {
	return userConfig.loggedIn()
}

// LogIn logs in using the given api key. This will delete the previously
// used api key and update the user's config.
func LogIn(apiKey string) error {
	if !ValidApiKey(apiKey) {
		return ErrInvalidAPIKey
	}
	if LoggedIn() && !Yes {
		fmt.Fprintf(os.Stderr, "you are already logged in, this action will deauthenticate the existing user.\ncontinue? [y/n]")
		for {
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
				os.Exit(1)
			}
			if response[0] == 'n' {
				os.Exit(1)
			}
			if response[0] == 'y' {
				break
			}
		}
	}
	refreshToken := "1234qwer0987poiu"
	// TODO: actually get a new refresh token, error if bad credentials
	err := userConfig.changeAndWrite(config{
		ApiKey:       apiKey,
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

func ValidApiKey(apiKey string) bool {
	return len(apiKey) == 8 // TODO
}
