package logout

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/profile"
	"github.com/10gen/realm-cli/internal/terminal"
)

// Command is the `logout` command
var Command = cli.CommandDefinition{
	Use:         "logout",
	Description: "Terminate the current user’s session",
	Help:        "logout", // TODO(REALMC-7429): add help text description
	Command:     &command{},
}

type command struct{}

func (cmd *command) Handler(profile *profile.Profile, ui terminal.UI) error {
	user := profile.User()
	user.PrivateAPIKey = "" // ensures subsequent `login` commands prompt for password

	profile.SetUser(user.PublicAPIKey, user.PrivateAPIKey)
	profile.ClearSession()

	return profile.Save()
}

func (cmd *command) Feedback(profile *profile.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully logged out"))
}
