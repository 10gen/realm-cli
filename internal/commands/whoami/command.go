package whoami

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
)

// Command is the `whoami` command
var Command = cli.CommandDefinition{
	Use:         "whoami",
	Description: "Display the current user's details",
	Help:        "whoami", // TODO(REALMC-7429): add help text description
	Command:     &command{},
}

type command struct{}

func (cmd *command) Handler(profile *cli.Profile, ui terminal.UI) error {
	return nil // commands without handlers show help text and usage when ran
}

func (cmd *command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	user := profile.User()
	session := profile.Session()

	if user.PublicAPIKey == "" {
		return ui.Print(terminal.NewTextLog("No user is currently logged in"))
	}

	if session.AccessToken == "" {
		return ui.Print(terminal.NewTextLog("The user, %s, is not currently logged in", user.PublicAPIKey))
	}

	return ui.Print(terminal.NewTextLog("Currently logged in user: %s (%s)", user.PublicAPIKey, user.RedactedPrivateAPIKey()))
}
