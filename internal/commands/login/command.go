package login

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
	"github.com/10gen/realm-cli/internal/utils/flags"
)

// CommandMeta is the command meta for the `login` command
var CommandMeta = cli.CommandMeta{
	Use:         "login",
	Description: "Log the CLI into Realm using a MongoDB Cloud API key",
	HelpText: `Begins an authenticated session with Realm. To get a MongoDB Cloud API Key, open
your Realm app in the Realm UI. Navigate to "Deployment" in the left navigation
menu, and select the "Export App" tab. From there, create a programmatic API key
to authenticate your realm-cli session.`,
}

// Command is the `login` command
type Command struct {
	inputs inputs
}

// Flags is the command flags
func (cmd *Command) Flags() []flags.Flag {
	return []flags.Flag{
		flags.StringFlag{
			Value: &cmd.inputs.PublicAPIKey,
			Meta: flags.Meta{
				Name: "api-key",
				Usage: flags.Usage{
					Description: "Specify the public portion of your Atlas programmatic API Key",
				},
			},
		},
		flags.StringFlag{
			Value: &cmd.inputs.PrivateAPIKey,
			Meta: flags.Meta{
				Name: "private-api-key",
				Usage: flags.Usage{
					Description: "Specify the private portion of your Atlas programmatic API Key",
				},
			},
		},
	}
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
