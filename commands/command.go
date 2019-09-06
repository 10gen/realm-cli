// Package commands provides all commands associated with the CLI, and a means
// of executing them.
package commands

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/api/mdbcloud"
	"github.com/10gen/stitch-cli/storage"
	"github.com/10gen/stitch-cli/user"
	"github.com/10gen/stitch-cli/utils"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
)

const (
	flagAppIDName = "app-id"
)

var (
	errAppIDRequired = fmt.Errorf("an App ID (--%s=[string]) must be supplied to export an app", flagAppIDName)
)

// BaseCommand handles the parsing and execution of a command.
type BaseCommand struct {
	*flag.FlagSet

	Name string

	CLI *cli.CLI
	UI  cli.Ui

	client       api.Client
	atlasClient  mdbcloud.Client
	stitchClient api.StitchClient
	user         *user.User
	storage      *storage.Storage

	flagConfigPath    string
	flagColorDisabled bool
	flagBaseURL       string
	flagAtlasBaseURL  string
	flagYes           bool
}

// NewFlagSet builds and returns the default set of flags for all commands
func (c *BaseCommand) NewFlagSet() *flag.FlagSet {
	set := flag.NewFlagSet(c.Name, flag.ExitOnError)
	set.Usage = func() {}

	set.BoolVar(&c.flagColorDisabled, "disable-color", false, "")
	set.BoolVar(&c.flagYes, "yes", false, "")
	set.BoolVar(&c.flagYes, "y", false, "")
	set.StringVar(&c.flagBaseURL, "base-url", api.DefaultBaseURL, "")
	set.StringVar(&c.flagAtlasBaseURL, "atlas-base-url", api.DefaultAtlasBaseURL, "")
	set.StringVar(&c.flagConfigPath, "config-path", "", "")

	c.FlagSet = set

	return set
}

// Client returns an api.Client for use with API calls to services
func (c *BaseCommand) Client() (api.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	c.client = api.NewClient(c.flagBaseURL)

	return c.client, nil
}

// AtlasClient returns a mdbcloud.Client for use with MDB Cloud Manager APIs
func (c *BaseCommand) AtlasClient() (mdbcloud.Client, error) {
	if c.atlasClient != nil {
		return c.atlasClient, nil
	}

	user, err := c.User()
	if err != nil {
		return nil, err
	}

	c.atlasClient = mdbcloud.NewClient(c.flagAtlasBaseURL).WithAuth(user.PublicAPIKey, user.PrivateAPIKey)

	return c.atlasClient, nil
}

// AuthClient returns an api.Client that is aware of the current user's auth credentials. It also handles retrying
// requests if a user's access token has expired
func (c *BaseCommand) AuthClient() (api.Client, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	user, err := c.User()
	if err != nil {
		return nil, err
	}

	authClient := api.NewAuthClient(client, user)

	tokenIsExpired, err := user.TokenIsExpired()
	if err != nil {
		return nil, err
	}

	if tokenIsExpired {
		authResponse, err := authClient.RefreshAuth()
		if err != nil {
			return nil, err
		}

		user.AccessToken = authResponse.AccessToken

		if err := c.storage.WriteUserConfig(user); err != nil {
			return nil, err
		}
	}

	return authClient, nil
}

// StitchClient returns an api.StitchClient for use in calling the API
func (c *BaseCommand) StitchClient() (api.StitchClient, error) {
	if c.stitchClient != nil {
		return c.stitchClient, nil
	}

	authClient, err := c.AuthClient()
	if err != nil {
		return nil, err
	}

	c.stitchClient = api.NewStitchClient(authClient)

	return c.stitchClient, nil
}

// User returns the current user. It loads the user from storage if it is not available in memory
func (c *BaseCommand) User() (*user.User, error) {
	if c.user != nil {
		return c.user, nil
	}

	u, err := c.storage.ReadUserConfig()
	if err != nil {
		return nil, err
	}

	c.user = u

	return u, nil
}

func (c *BaseCommand) run(args []string) error {
	if c.FlagSet == nil {
		c.NewFlagSet()
	}

	// FlagSet uses flag.ExitOnError, so we let it handle flag-related errors
	// to avoid duplicate error output
	c.Parse(args)

	if !c.flagColorDisabled && isatty.IsTerminal(os.Stdout.Fd()) {
		c.UI = &cli.ColoredUi{
			ErrorColor: cli.UiColorRed,
			WarnColor:  cli.UiColorYellow,
			Ui:         c.UI,
		}
	}

	if url := utils.CheckForNewCLIVersion(http.DefaultClient); url != "" {
		c.UI.Info(url)
	}

	if c.storage == nil {
		path, err := homedir.Expand(c.flagConfigPath)
		if err != nil {
			return err
		}

		if path == "" {
			home, dirErr := homedir.Dir()
			if dirErr != nil {
				return dirErr
			}
			path = filepath.Join(home, ".config", "stitch", "stitch")
		}

		fileStrategy, err := storage.NewFileStrategy(path)
		if err != nil {
			return err
		}

		c.storage = storage.New(fileStrategy)
	}

	return nil
}

// AskYesNo is used to prompt the user for yes/no input
func (c *BaseCommand) AskYesNo(query string) (bool, error) {
	if c.flagYes {
		c.UI.Info(fmt.Sprintf("%s [y/n]: y", query))
		return true, nil
	}

	res, err := c.UI.Ask(query + " [y/n]:")
	if err != nil {
		return false, err
	}

	for {
		var answer string

		if len(res) > 0 {
			answer = strings.TrimSpace(strings.ToLower(res))
		}

		if nay(answer) {
			return false, nil
		}

		if yay(answer) {
			return true, nil
		}

		res, _ = c.UI.Ask("Could not understand response, try again [y/n]:")
	}
}

// Ask is used to prompt the user for input
func (c *BaseCommand) Ask(query string, defaultVal string) (string, error) {
	if c.flagYes && defaultVal != "" {
		c.UI.Info(fmt.Sprintf("%s [%s]: %s", query, defaultVal, defaultVal))
		return defaultVal, nil
	}

	var defaultClause string
	if defaultVal != "" {
		defaultClause = fmt.Sprintf(" [%s]", defaultVal)
	}
	res, err := c.UI.Ask(fmt.Sprintf("%s%s:", query, defaultClause))
	if err != nil {
		return "", err
	}

	for {
		var answer string

		if len(res) > 0 {
			answer = strings.TrimSpace(res)
		}

		if answer == "" && defaultVal != "" {
			return defaultVal, nil
		}

		if len(answer) != 0 {
			return answer, nil
		}

		res, _ = c.UI.Ask(fmt.Sprintf("Could not understand response, try again%s:", defaultClause))
	}
}

// AskWithOptions is used to prompt user for input from a list of options
func (c *BaseCommand) AskWithOptions(query, defaultValue string, options []string) (string, error) {
	if c.flagYes && defaultValue != "" {
		c.UI.Info(fmt.Sprintf("%s [%s]: %s", query, defaultValue, defaultValue))
		return defaultValue, nil
	}

	var defaultClause string
	if defaultValue != "" {
		defaultClause = fmt.Sprintf(" [%s]", defaultValue)
	}
	res, err := c.UI.Ask(fmt.Sprintf("%s%s:", query, defaultClause))
	if err != nil {
		return "", err
	}

	for {
		answer := strings.TrimSpace(res)

		if answer == "" && defaultValue != "" {
			return defaultValue, nil
		}

		if answer != "" {
			for _, option := range options {
				if strings.EqualFold(answer, option) {
					return option, nil
				}
			}
		}

		res, _ = c.UI.Ask(fmt.Sprintf("Could not understand response, valid values are %s:", strings.Join(options, ", ")))
	}
}

// Help defines help documentation for parameters that apply to all commands
func (c *BaseCommand) Help() string {
	return `

  --config-path [string]
	File to write user configuration data to (defaults to ~/.config/stitch/stitch)

  --disable-color
	Disable the use of colors in terminal output.

  -y, --yes
	Bypass prompts. Provide this parameter if you do not want to be prompted for input.`
}

func yay(s string) bool {
	return s == "y" || s == "yes"
}

func nay(s string) bool {
	return s == "n" || s == "no"
}
