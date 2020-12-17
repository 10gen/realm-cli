package login

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is the `login` command
var Command = cli.CommandDefinition{
	Use:         "login",
	Description: "Authenticate with an Atlas programmatic API Key",
	Help:        "login", // TODO(REALMC-7429): add help text description
	Command:     &command{},
}

type command struct {
	inputs      inputs
	realmClient realm.Client
}

func (cmd *command) Flags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.PublicAPIKey, flagPublicAPIKey, flagPublicAPIKeyShort, "", flagPublicAPIKeyUsage)
	fs.StringVarP(&cmd.inputs.PrivateAPIKey, flagPrivateAPIKey, flagPrivateAPIKeyShort, "", flagPrivateAPIKeyUsage)
}

func (cmd *command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

func (cmd *command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewClient(profile.RealmBaseURL())
	return nil
}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	existingUser := profile.User()

	if existingUser.PublicAPIKey != "" && existingUser.PublicAPIKey != cmd.inputs.PublicAPIKey {
		proceed, err := ui.Confirm(
			"This action will terminate the existing session for user: %s (%s), would you like to proceed?",
			existingUser.PublicAPIKey,
			existingUser.RedactedPrivateAPIKey(),
		)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
	}

	profile.SetUser(cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey)

	auth, authErr := cmd.realmClient.Authenticate(cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey)
	if authErr != nil {
		return authErr
	}

	profile.SetSession(auth.AccessToken, auth.RefreshToken)
	return profile.Save()
}

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully logged in"))
}
