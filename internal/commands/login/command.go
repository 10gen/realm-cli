package login

import (
	"github.com/10gen/realm-cli/internal/auth"
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cloud/realm"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is the `login` command
type Command struct {
	inputs      inputs
	realmClient realm.Client
}

// Flags is the command flags
func (cmd *Command) Flags(fs *pflag.FlagSet) {
	fs.StringVarP(&cmd.inputs.PublicAPIKey, flagPublicAPIKey, flagPublicAPIKeyShort, "", flagPublicAPIKeyUsage)
	fs.StringVarP(&cmd.inputs.PrivateAPIKey, flagPrivateAPIKey, flagPrivateAPIKeyShort, "", flagPrivateAPIKeyUsage)
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Setup is the command setup
func (cmd *Command) Setup(profile *cli.Profile, ui terminal.UI) error {
	cmd.realmClient = realm.NewClient(profile.RealmBaseURL())
	return nil
}

// Handler is the command handler
func (cmd *Command) Handler(profile *cli.Profile, ui terminal.UI) error {
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

	profile.SetUser(auth.User{cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey})

	session, sessionErr := cmd.realmClient.Authenticate(cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey)
	if sessionErr != nil {
		return sessionErr
	}

	profile.SetSession(auth.Session{session.AccessToken, session.RefreshToken})
	return profile.Save()
}

// Feedback is the command feedback
func (cmd *Command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully logged in"))
}
