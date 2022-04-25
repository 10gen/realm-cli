package logout

import (
	"github.com/10gen/realm-cli/internal/cli"
	"github.com/10gen/realm-cli/internal/cli/user"
	"github.com/10gen/realm-cli/internal/terminal"
)

// CommandMeta is the command meta for the `logout` command
var CommandMeta = cli.CommandMeta{
	Use:         "logout",
	Description: "Log the CLI out of Realm",
	HelpText: `Ends the authenticated session and deletes cached auth tokens. To
re-authenticate, you must call Login with your Atlas programmatic API Key.`,
}

// Command is the `logout` command
type Command struct{}

// Handler is the command handler
func (cmd *Command) Handler(profile *user.Profile, ui terminal.UI, clients cli.Clients) error {
	profile.ClearCredentials()
	profile.ClearSession()

	if err := profile.Save(); err != nil {
		return err
	}

	ui.Print(terminal.NewTextLog("Successfully logged out"))
	return nil
}
