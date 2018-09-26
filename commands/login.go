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

	flagAPIKey       string
	flagUsername     string
	flagAuthProvider string
	flagPassword     string
}

// Synopsis returns a one-liner description for this command
func (lc *LoginCommand) Synopsis() string {
	return `Log in using an Atlas API Key`
}

// Help returns long-form help information for this command
func (lc *LoginCommand) Help() string {
	return `Authenticate as an administrator.

REQUIRED:
  --api-key [string]
	The API key for a MongoDB Cloud account.

  --username [string]
	The username for a MongoDB Cloud account.

OPTIONS:` +
		lc.BaseCommand.Help()
}

// Run executes the command
func (lc *LoginCommand) Run(args []string) int {
	set := lc.NewFlagSet()

	set.StringVar(&lc.flagAPIKey, flagLoginAPIKeyName, "", "")
	set.StringVar(&lc.flagAuthProvider, "auth-provider", string(auth.ProviderTypeAPIKey), "")
	set.StringVar(&lc.flagPassword, "password", "", "")
	set.StringVar(&lc.flagUsername, flagLoginUsernameName, "", "")

	if err := lc.BaseCommand.run(args); err != nil {
		lc.UI.Error(err.Error())
		return 1
	}

	if err := lc.logIn(); err != nil {
		lc.UI.Error(err.Error())
		return 1
	}

	lc.UI.Info(fmt.Sprintf("you have successfully logged in as %s", lc.flagUsername))
	return 0
}

func (lc *LoginCommand) validateAuthCredentials() (auth.AuthenticationProvider, error) {
	var provider auth.AuthenticationProvider

	switch auth.ProviderType(lc.flagAuthProvider) {
	case auth.ProviderTypeAPIKey:
		provider = auth.NewAPIKeyProvider(lc.flagUsername, lc.flagAPIKey)
	case auth.ProviderTypeUsernamePassword:
		if lc.flagPassword == "" {
			password, err := lc.UI.AskSecret("Password:")
			if err != nil {
				return nil, err
			}

			lc.flagPassword = password
		}
		provider = auth.NewUsernamePasswordProvider(lc.flagUsername, lc.flagPassword)
	default:
		return nil, fmt.Errorf("invalid authentication provider")
	}

	if err := provider.Validate(); err != nil {
		return nil, err
	}

	return provider, nil
}

func (lc *LoginCommand) logIn() error {
	authProvider, err := lc.validateAuthCredentials()
	if err != nil {
		return err
	}

	user, err := lc.User()
	if err != nil {
		return err
	}

	if user.LoggedIn() {
		shouldContinue, askErr := lc.AskYesNo(fmt.Sprintf(
			"you are already logged in as %s, this action will deauthenticate the existing user [apiKey: %s].\ncontinue?",
			user.Username,
			user.RedactedAPIKey(),
		))

		if askErr != nil {
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

	authResponse, err := api.NewStitchClient(client).Authenticate(authProvider)
	if err != nil {
		return err
	}

	user.APIKey = lc.flagAPIKey
	user.Username = lc.flagUsername
	user.AccessToken = authResponse.AccessToken
	user.RefreshToken = authResponse.RefreshToken

	return lc.storage.WriteUserConfig(user)
}
