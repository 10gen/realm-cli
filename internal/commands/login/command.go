package login

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"

	"github.com/spf13/pflag"
)

// Command is the `login` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags(fs *pflag.FlagSet) {
	fs.StringVar(&cmd.inputs.PublicAPIKey, flagPublicAPIKey, "", flagPublicAPIKeyUsage)
	fs.StringVar(&cmd.inputs.PrivateAPIKey, flagPrivateAPIKey, "", flagPrivateAPIKeyUsage)
}

// Inputs is the command inputs
func (cmd *Command) Inputs() cli.InputResolver {
	return &cmd.inputs
}

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	existingUser := profile.Credentials()

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

	profile.SetCredentials(user.Credentials{cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey})

	session, err := clients.Realm.Authenticate(cmd.inputs.PublicAPIKey, cmd.inputs.PrivateAPIKey)
	if err != nil {
		return err
	}

	profile.SetSession(user.Session{session.AccessToken, session.RefreshToken})
	if err := profile.Save(); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully logged in"))
	return nil
}
