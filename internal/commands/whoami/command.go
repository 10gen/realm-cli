package whoami

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
)

// Command is the `whoami` command
type Command struct{}

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	user := profile.Credentials()
	session := profile.Session()

	if user.PrivateAPIKey == "" {
		ui.Print(terminal.NewTextLog("No user is currently logged in"))
		return nil
	}

	if session.AccessToken == "" {
		ui.Print(terminal.NewTextLog("The user, %s, is not currently logged in", user.PublicAPIKey))
		return nil
	}

	ui.Print(terminal.NewTextLog("Currently logged in user: %s (%s)", user.PublicAPIKey, user.RedactedPrivateAPIKey()))
	return nil
}
