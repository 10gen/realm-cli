package logout

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/terminal"
)

// Command is the `logout` command
type Command struct{}

// Handler is the command handler
func (cmd *Command) Handler(profile *cli.Profile, ui terminal.UI) error {
	user := profile.User()
	user.PrivateAPIKey = "" // ensures subsequent `login` commands prompt for password

	profile.SetUser(user)
	profile.ClearSession()

	return profile.Save()
}

// Feedback is the command feedback
func (cmd *Command) Feedback(profile *cli.Profile, ui terminal.UI) error {
	return ui.Print(terminal.NewTextLog("Successfully logged out"))
}
