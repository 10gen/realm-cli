package cli

import (
	"fmt"

	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/flags"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/AlecAivazis/survey/v2"

	"github.com/spf13/pflag"
)

// LoginCommand creates the 'login' command
func LoginCommand() CommandDefinition {
	return CommandDefinition{
		Command:     &loginCommand{},
		Use:         "login",
		Description: "Authenticate with an Atlas programmatic API Key",
		Help:        "login", // TODO(REALMC-7429): add help text description
	}
}

type loginCommand struct {
	publicAPIKey  string
	privateAPIKey string

	realmClient realm.Client
}

func (cmd *loginCommand) RegisterFlags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.publicAPIKey, flags.PublicAPIKey, "", flags.PublicAPIKeyUsage)
	fs.StringVar(&cmd.privateAPIKey, flags.PrivateAPIKey, "", flags.PrivateAPIKeyUsage)
}

func (cmd *loginCommand) Setup(profile *Profile, ui terminal.UI, config CommandConfig) error {
	cmd.realmClient = realm.NewClient(config.RealmBaseURL)

	if cmd.publicAPIKey == "" {
		err := ui.AskOne(
			&survey.Input{Message: "API Key", Default: profile.GetUser().PublicAPIKey},
			&cmd.publicAPIKey,
		)
		if err != nil {
			return err
		}
	}

	if cmd.privateAPIKey == "" {
		err := ui.AskOne(
			&survey.Password{Message: "Private API Key"},
			&cmd.privateAPIKey,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cmd *loginCommand) Handler(profile *Profile, ui terminal.UI, args []string) error {
	existingUser := profile.GetUser()

	if existingUser.PublicAPIKey != "" && existingUser.PublicAPIKey != cmd.publicAPIKey {
		var proceed bool

		err := ui.AskOne(
			&survey.Confirm{Message: fmt.Sprintf(
				"This action will terminate the existing session for user: %s (%s), would you like to proceed?",
				existingUser.PublicAPIKey,
				existingUser.RedactedPrivateAPIKey(),
			)},
			&proceed)
		if err != nil {
			return err
		}

		if !proceed {
			return nil
		}
	}

	profile.SetUser(cmd.publicAPIKey, cmd.privateAPIKey)

	auth, authErr := cmd.realmClient.Authenticate(cmd.publicAPIKey, cmd.privateAPIKey)
	if authErr != nil {
		return authErr
	}

	profile.SetSession(auth.AccessToken, auth.RefreshToken)
	return profile.Save()
}

func (cmd *loginCommand) Feedback(profile *Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully logged in."))
}
