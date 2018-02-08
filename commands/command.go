// Package commands provides all commands associated with the CLI, and a means
// of executing them.
package commands

import (
	"path/filepath"
	"strings"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/storage"
	"github.com/10gen/stitch-cli/user"

	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
	flag "github.com/ogier/pflag"
)

// BaseCommand handles the parsing and execution of a command.
type BaseCommand struct {
	*flag.FlagSet

	Name string

	UI cli.Ui

	client  api.Client
	user    *user.User
	storage *storage.Storage

	flagConfigPath   string
	flagColorEnabled bool
	flagBaseURL      string
	flagYes          bool
}

// NewFlagSet builds and returns the default set of flags for all commands
func (c *BaseCommand) NewFlagSet() *flag.FlagSet {
	set := flag.NewFlagSet(c.Name, flag.ContinueOnError)
	set.SetInterspersed(true)
	set.Usage = func() {}

	set.BoolVar(&c.flagColorEnabled, "color", true, "")
	set.BoolVarP(&c.flagYes, "yes", "y", false, "")
	set.StringVar(&c.flagBaseURL, "base-url", api.DefaultBaseURL, "")
	set.StringVar(&c.flagConfigPath, "config-path", "", "")

	c.FlagSet = set

	return set
}

// Client returns an api.Client for use with API calls to services
func (c *BaseCommand) Client() (api.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	c.client = api.NewClient()

	return c.client, nil
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

	authClient := api.NewAuthClient(c.flagBaseURL, client, user)

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

	if err := c.Parse(args); err != nil {
		return err
	}

	if c.storage == nil {
		path, err := homedir.Expand(c.flagConfigPath)
		if err != nil {
			return err
		}

		if path == "" {
			home, err := homedir.Dir()
			if err != nil {
				return err
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

// Ask is used to prompt the user for input
func (c *BaseCommand) Ask(query string) (bool, error) {
	if c.flagYes {
		return true, nil
	}

	res, err := c.UI.Ask(query + " [y/n] ")
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

		res, _ = c.UI.Ask("Could not understand response, try again [y/n]: ")
	}
}

func yay(s string) bool {
	return s == "y" || s == "yes"
}

func nay(s string) bool {
	return s == "n" || s == "no"
}
