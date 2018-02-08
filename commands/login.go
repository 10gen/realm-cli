package commands

import (
	"fmt"

	"github.com/10gen/stitch-cli/api"
	"github.com/10gen/stitch-cli/auth"

	"github.com/mitchellh/cli"
)

const (
	flagLoginAPIKeyName   = "api-key"
	flagLoginUsernameName = "username"
)

var (
	errAPIKeyRequired   = fmt.Errorf("an API key (--%s=<TOKEN>) must be supplied to log in", flagLoginAPIKeyName)
	errUsernameRequired = fmt.Errorf("a username (--%s=<USERNAME>) must be supplied to log in", flagLoginUsernameName)
)

// NewLoginCommandFactory returns a new cli.CommandFactory given a cli.Ui
func NewLoginCommandFactory(ui cli.Ui) cli.CommandFactory {
	return func() (cli.Command, error) {
		return &LoginCommand{
			BaseCommand: &BaseCommand{
				Name: "login",
				UI:   ui,
			},
		}, nil
	}
}

// LoginCommand is used to authenticate a user given an API key and username
type LoginCommand struct {
	*BaseCommand

	flagAPIKey   string
	flagUsername string
}

// Synopsis returns a one-liner description for this command
func (lc *LoginCommand) Synopsis() string {
	return `Log in using an Atlas API Key`
}

// Help returns long-form help information for this command
func (lc *LoginCommand) Help() string {
	return `Authenticate as an administrator.

OPTIONS:
  --api-key <TOKEN>
	The API key for a MongoDB Cloud account.
`
}

// Run executes the command
func (lc *LoginCommand) Run(args []string) int {
	set := lc.NewFlagSet()

	set.StringVar(&lc.flagAPIKey, flagLoginAPIKeyName, "", "")
	set.StringVar(&lc.flagUsername, flagLoginUsernameName, "", "")

	if err := lc.BaseCommand.run(args); err != nil {
		lc.UI.Error(err.Error())
		return 1
	}

	if err := lc.logIn(); err != nil {
		lc.UI.Error(err.Error())
		return 1
	}

	return 0
}

func (lc *LoginCommand) logIn() error {
	if lc.flagAPIKey == "" {
		return errAPIKeyRequired
	}

	if lc.flagUsername == "" {
		return errUsernameRequired
	}

	if !auth.ValidAPIKey(lc.flagAPIKey) {
		return auth.ErrInvalidAPIKey
	}

	user, err := lc.User()
	if err != nil {
		return err
	}

	if user.LoggedIn() {
		shouldContinue, err := lc.Ask("you are already logged in, this action will deauthenticate the existing user.\ncontinue?")
		if err != nil {
			return err
		}

		if !shouldContinue {
			return nil
		}
	}

	client, err := lc.Client()
	if err != nil {
		return err
	}

	authResponse, err := api.NewStitchClient(lc.flagBaseURL, client).Authenticate(lc.flagAPIKey, lc.flagUsername)
	if err != nil {
		return err
	}

	user.APIKey = lc.flagAPIKey
	user.Username = lc.flagUsername
	user.AccessToken = authResponse.AccessToken
	user.RefreshToken = authResponse.RefreshToken

	return lc.storage.WriteUserConfig(user)
}
